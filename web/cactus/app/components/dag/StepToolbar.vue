<script setup lang="ts">
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap,
  PanelLeftClose, PanelLeftOpen, GripVertical, Search,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import type { WorkTypeCatalogItem } from '~/composables/useWorkers'
import type { ControlStepDefinition } from '~/composables/useControlSteps'
import { filterStepCatalog } from '~/components/dag/step-toolbar-utils'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'

const props = withDefaults(defineProps<{
  disabled?: boolean
}>(), {
  disabled: false,
})

const emit = defineEmits<{
  addStep: [
    stepType: string,
    workTypeId: number | undefined,
    workTypeCode: string | undefined,
    workerSettingsSchemaId: number | undefined,
    position: { x: number, y: number },
    name: string | undefined,
  ]
}>()

const { t } = useI18n()
const { fetchWorkTypeCatalog } = useWorkers()
const controlSteps = useControlSteps()

const workTypes = ref<WorkTypeCatalogItem[]>([])
const collapsed = ref(false)
const search = ref('')
const schemaDialogOpen = ref(false)
const pendingSchemaItem = ref<ToolbarItem | null>(null)
const selectedSchemaByWorkType = ref<Record<number, number>>({})

const iconMap: Record<string, Component> = {
  'mail': Mail,
  'message-square': MessageSquare,
  'bell': Bell,
  'workflow': Workflow,
  'git-branch': GitBranch,
  'clock': Clock,
  'split': Split,
  'zap': Zap,
}

interface ToolbarItem {
  id: string
  name: string
  icon: Component
  stepType: 'task' | 'control'
  workTypeId?: number
  workTypeCode?: string
  controlKind?: string
  category: string
  available: boolean
  disabledReason?: string
  readyWorkers?: number
  workerCount?: number
  schemas?: Array<{ id: number, version: string }>
}

const taskItems = computed<ToolbarItem[]>(() =>
  workTypes.value.map((wt) => {
    const meta = wt.meta
    const iconName = meta?.icon ?? 'workflow'
    const available = wt.ready_workers > 0

    return {
      id: `wt-${wt.id}`,
      name: wt.name,
      icon: iconMap[iconName] ?? Workflow,
      stepType: 'task',
      workTypeId: wt.id,
      workTypeCode: wt.code,
      category: meta?.category ?? 'Other',
      available,
      disabledReason: available ? undefined : t('toolbar.noWorkersAvailable'),
      readyWorkers: wt.ready_workers,
      workerCount: wt.worker_count,
      schemas: wt.schemas,
    }
  }),
)

const controlItems = computed<ToolbarItem[]>(() =>
  controlSteps.map((control: ControlStepDefinition) => ({
    id: `control-${control.kind}`,
    name: control.name,
    icon: iconMap[control.icon] ?? GitBranch,
    stepType: 'control',
    workTypeCode: control.kind,
    controlKind: control.kind,
    category: control.category,
    available: true,
  })),
)

const filteredItems = computed(() =>
  filterStepCatalog([...taskItems.value, ...controlItems.value], search.value),
)

const groupedItems = computed(() => {
  const groups: Record<string, ToolbarItem[]> = {}
  for (const item of filteredItems.value) {
    if (!groups[item.category]) {
      groups[item.category] = []
    }
    groups[item.category].push(item)
  }
  return groups
})

const pendingSchemas = computed(() => pendingSchemaItem.value?.schemas ?? [])

function selectedSchemaId(item: ToolbarItem) {
  if (!item.workTypeId || !item.schemas?.length) return undefined
  return selectedSchemaByWorkType.value[item.workTypeId] ?? item.schemas[0]?.id
}

function requiresSchemaChoice(item: ToolbarItem) {
  return item.stepType === 'task' && (item.schemas?.length ?? 0) > 1 && item.workTypeId && !selectedSchemaByWorkType.value[item.workTypeId]
}

function openSchemaDialog(item: ToolbarItem) {
  pendingSchemaItem.value = item
  schemaDialogOpen.value = true
}

function chooseSchema(schemaId: number) {
  const item = pendingSchemaItem.value
  if (!item?.workTypeId) return
  selectedSchemaByWorkType.value = {
    ...selectedSchemaByWorkType.value,
    [item.workTypeId]: schemaId,
  }
  schemaDialogOpen.value = false
  pendingSchemaItem.value = null
}

function payloadFor(item: ToolbarItem) {
  return {
    stepType: item.stepType,
    workTypeId: item.workTypeId,
    workTypeCode: item.stepType === 'control' ? item.controlKind : item.workTypeCode,
    workerSettingsSchemaId: item.stepType === 'task' ? selectedSchemaId(item) : undefined,
    name: item.name,
  }
}

function onDragStart(event: DragEvent, item: ToolbarItem) {
  if (props.disabled || !item.available) {
    event.preventDefault()
    return
  }
  if (requiresSchemaChoice(item)) {
    event.preventDefault()
    openSchemaDialog(item)
    return
  }
  if (!event.dataTransfer) return

  event.dataTransfer.setData('application/cactus-step', JSON.stringify(payloadFor(item)))
  event.dataTransfer.effectAllowed = 'move'
}

function onDoubleClick(item: ToolbarItem) {
  if (props.disabled || !item.available) return
  if (requiresSchemaChoice(item)) {
    openSchemaDialog(item)
    return
  }
  const payload = payloadFor(item)
  emit('addStep', payload.stepType, payload.workTypeId, payload.workTypeCode, payload.workerSettingsSchemaId, { x: 300, y: 200 }, payload.name)
}

onMounted(async () => {
  try {
    workTypes.value = await fetchWorkTypeCatalog()
  }
  catch {
    workTypes.value = []
  }
})
</script>

<template>
  <div
    class="flex flex-col border-r bg-background transition-all duration-200"
    :class="[collapsed ? 'w-12' : 'w-72', disabled ? 'opacity-60' : '']"
  >
    <div class="flex items-center justify-between border-b p-2">
      <span v-if="!collapsed" class="text-sm font-semibold">
        {{ t('editor.steps') }}
      </span>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="collapsed = !collapsed"
      >
        <PanelLeftClose v-if="!collapsed" class="h-4 w-4" />
        <PanelLeftOpen v-else class="h-4 w-4" />
      </Button>
    </div>

    <div v-if="!collapsed" class="border-b p-2">
      <div class="relative">
        <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          v-model="search"
          class="h-9 pl-8"
          :placeholder="t('toolbar.searchSteps')"
        />
      </div>
    </div>

    <ScrollArea v-if="!collapsed" class="flex-1">
      <div class="space-y-4 p-2">
        <div v-for="(groupItems, category) in groupedItems" :key="category" class="space-y-1">
          <p class="px-2 text-xs font-medium text-muted-foreground">
            {{ category }}
          </p>
          <div
            v-for="item in groupItems"
            :key="item.id"
            class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
            :class="item.available && !disabled ? 'cursor-grab hover:bg-accent active:cursor-grabbing' : 'cursor-not-allowed bg-muted/40 text-muted-foreground'"
            :draggable="item.available && !disabled"
            :title="item.disabledReason"
            @dragstart="onDragStart($event, item)"
            @dblclick="onDoubleClick(item)"
          >
            <GripVertical class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <component :is="item.icon" class="h-4 w-4 shrink-0" />
            <div class="min-w-0 flex-1">
              <p class="truncate">{{ item.name }}</p>
              <p v-if="item.stepType === 'task'" class="text-xs text-muted-foreground">
                {{ t('toolbar.readyWorkers', { ready: item.readyWorkers ?? 0, total: item.workerCount ?? 0 }) }}
              </p>
            </div>
            <Badge variant="secondary" class="h-5 px-1.5 text-[10px]">
              {{ item.stepType }}
            </Badge>
          </div>
        </div>

        <p v-if="filteredItems.length === 0" class="py-4 text-center text-xs text-muted-foreground">
          {{ t('common.noResults') }}
        </p>
      </div>
    </ScrollArea>

    <Dialog v-model:open="schemaDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('toolbar.chooseSchema') }}</DialogTitle>
          <DialogDescription>{{ pendingSchemaItem?.name }}</DialogDescription>
        </DialogHeader>
        <div class="space-y-2">
          <Button
            v-for="schema in pendingSchemas"
            :key="schema.id"
            variant="outline"
            class="w-full justify-between"
            @click="chooseSchema(schema.id)"
          >
            <span>{{ t('toolbar.schemaVersion', { version: schema.version }) }}</span>
            <span class="text-xs text-muted-foreground">#{{ schema.id }}</span>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

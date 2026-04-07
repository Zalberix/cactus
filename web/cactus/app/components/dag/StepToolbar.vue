<script setup lang="ts">
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap,
  PanelLeftClose, PanelLeftOpen, GripVertical,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import type { WorkType } from '~/composables/useWorkers'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Button } from '~/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '~/components/ui/collapsible'

const emit = defineEmits<{
  addStep: [stepType: string, workTypeId: number | undefined, position: { x: number; y: number }]
}>()

const { t } = useI18n()
const { fetchWorkTypes } = useVersions()

const workTypes = ref<WorkType[]>([])
const collapsed = ref(false)

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
  stepType: string
  workTypeId?: number
  workTypeCode?: string
  category: string
}

const items = computed<ToolbarItem[]>(() => {
  const list: ToolbarItem[] = []

  for (const wt of workTypes.value) {
    const meta = wt.meta
    const iconName = meta?.icon ?? 'workflow'
    const category = meta?.category ?? 'Other'
    const isControl = meta?.kind === 'control'
    const stepType = isControl ? 'control' : 'task'

    list.push({
      id: `wt-${wt.id}`,
      name: wt.name,
      icon: iconMap[iconName] ?? Workflow,
      stepType,
      workTypeId: isControl ? undefined : wt.id,
      workTypeCode: wt.code,
      category,
    })
  }

  return list
})

const groupedItems = computed(() => {
  const groups: Record<string, ToolbarItem[]> = {}
  for (const item of items.value) {
    if (!groups[item.category]) {
      groups[item.category] = []
    }
    groups[item.category].push(item)
  }
  return groups
})

function onDragStart(event: DragEvent, item: ToolbarItem) {
  if (!event.dataTransfer) return
  event.dataTransfer.setData('application/cactus-step', JSON.stringify({
    stepType: item.stepType,
    workTypeId: item.workTypeId,
    workTypeCode: item.workTypeCode,
    name: item.name,
  }))
  event.dataTransfer.effectAllowed = 'move'
}

function onDoubleClick(item: ToolbarItem) {
  emit('addStep', item.stepType, item.workTypeId, { x: 300, y: 200 })
}

onMounted(async () => {
  try {
    workTypes.value = await fetchWorkTypes()
  }
  catch {
    // Non-critical: items still available
  }
})
</script>

<template>
  <div
    class="flex flex-col border-r bg-background transition-all duration-200"
    :class="collapsed ? 'w-12' : 'w-60'"
  >
    <!-- Header -->
    <div class="flex items-center justify-between border-b p-2">
      <span
        v-if="!collapsed"
        class="text-sm font-semibold"
      >
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

    <!-- Items grouped by category -->
    <ScrollArea v-if="!collapsed" class="flex-1">
      <div class="p-2 space-y-1">
        <Collapsible
          v-for="(groupItems, category) in groupedItems"
          :key="category"
          :default-open="true"
        >
          <CollapsibleTrigger class="flex w-full items-center gap-1 rounded px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-accent">
            {{ category }}
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="mt-1 space-y-1">
              <div
                v-for="item in groupItems"
                :key="item.id"
                class="flex cursor-grab items-center gap-2 rounded-md border px-3 py-2 text-sm hover:bg-accent active:cursor-grabbing"
                draggable="true"
                @dragstart="onDragStart($event, item)"
                @dblclick="onDoubleClick(item)"
              >
                <GripVertical class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                <component :is="item.icon" class="h-4 w-4 shrink-0" />
                <span class="truncate">{{ item.name }}</span>
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>

        <p
          v-if="items.length === 0"
          class="py-4 text-center text-xs text-muted-foreground"
        >
          {{ t('common.loading') }}
        </p>
      </div>
    </ScrollArea>
  </div>
</template>

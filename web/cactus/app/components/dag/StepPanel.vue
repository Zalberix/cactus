<script setup lang="ts">
import { X, Trash2, Plus } from 'lucide-vue-next'
import type { Node } from '@vue-flow/core'
import type { StepData } from '~/composables/useDagEditor'
import type { WorkType } from '~/composables/useVersions'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import { Badge } from '~/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'

const props = defineProps<{
  node: Node
  workTypes: WorkType[]
}>()

const emit = defineEmits<{
  close: []
  updateData: [nodeId: string, data: Partial<StepData>]
  updateStep: [stepId: string, data: Record<string, unknown>]
  deleteStep: [stepId: string]
}>()

const { t } = useI18n()

const stepName = ref(props.node.data.label as string)
const configText = ref(
  props.node.data.config
    ? JSON.stringify(props.node.data.config, null, 2)
    : '{}',
)

// Input mapping as key-value pairs for editing
const mappingPairs = ref<Array<{ key: string; value: string }>>(
  Object.entries((props.node.data.inputMapping ?? {}) as Record<string, string>).map(
    ([key, value]) => ({ key, value }),
  ),
)

// Watch for node changes (when selecting different nodes)
watch(
  () => props.node.id,
  () => {
    stepName.value = props.node.data.label as string
    configText.value = props.node.data.config
      ? JSON.stringify(props.node.data.config, null, 2)
      : '{}'
    mappingPairs.value = Object.entries(
      (props.node.data.inputMapping ?? {}) as Record<string, string>,
    ).map(([key, value]) => ({ key, value }))
  },
)

function onNameBlur() {
  if (stepName.value === props.node.data.label) return
  emit('updateData', props.node.id, { label: stepName.value })
  emit('updateStep', props.node.id, { name: stepName.value })
}

function onWorkTypeChange(workTypeId: string) {
  const wt = props.workTypes.find(w => String(w.id) === workTypeId)
  emit('updateData', props.node.id, {
    workTypeId: wt?.id,
    workTypeName: wt?.name,
  })
  emit('updateStep', props.node.id, { work_type_id: wt?.id })
}

function onConfigBlur() {
  try {
    const parsed = JSON.parse(configText.value)
    emit('updateData', props.node.id, { config: parsed })
    emit('updateStep', props.node.id, { config: parsed })
  }
  catch {
    // Invalid JSON -- skip update
  }
}

function addMappingPair() {
  mappingPairs.value.push({ key: '', value: '' })
}

function removeMappingPair(index: number) {
  mappingPairs.value.splice(index, 1)
  syncMappings()
}

function syncMappings() {
  const mapping: Record<string, string> = {}
  for (const pair of mappingPairs.value) {
    if (pair.key.trim()) {
      mapping[pair.key.trim()] = pair.value
    }
  }
  emit('updateData', props.node.id, { inputMapping: mapping })
  emit('updateStep', props.node.id, { input_mapping: mapping })
}

function onDelete() {
  emit('deleteStep', props.node.id)
}
</script>

<template>
  <div class="flex w-[360px] flex-col border-l bg-background">
    <!-- Header -->
    <div class="flex items-center justify-between border-b px-4 py-3">
      <Input
        v-model="stepName"
        class="h-8 border-0 p-0 text-sm font-semibold shadow-none focus-visible:ring-0"
        @blur="onNameBlur"
        @keydown.enter="($event.target as HTMLInputElement)?.blur()"
      />
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7 shrink-0"
        @click="emit('close')"
      >
        <X class="h-4 w-4" />
      </Button>
    </div>

    <ScrollArea class="flex-1">
      <div class="space-y-4 p-4">
        <!-- Type -->
        <div>
          <Label class="text-xs text-muted-foreground">{{ t('editor.stepType') }}</Label>
          <div class="mt-1">
            <Badge variant="secondary">
              {{ node.data.stepType === 'control' ? t('editor.control') : t('editor.task') }}
            </Badge>
          </div>
        </div>

        <Separator />

        <!-- Work Type -->
        <div v-if="node.data.stepType === 'task'">
          <Label class="text-xs text-muted-foreground">{{ t('editor.workType') }}</Label>
          <Select
            :model-value="node.data.workTypeId ? String(node.data.workTypeId) : undefined"
            @update:model-value="onWorkTypeChange"
          >
            <SelectTrigger class="mt-1">
              <SelectValue placeholder="Select work type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="wt in workTypes"
                :key="wt.id"
                :value="String(wt.id)"
              >
                {{ wt.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Separator />

        <!-- Input Mapping -->
        <div>
          <div class="flex items-center justify-between">
            <Label class="text-xs text-muted-foreground">{{ t('editor.inputMapping') }}</Label>
            <Button
              variant="ghost"
              size="sm"
              class="h-6 px-2 text-xs"
              @click="addMappingPair"
            >
              <Plus class="mr-1 h-3 w-3" />
              {{ t('editor.addMapping') }}
            </Button>
          </div>
          <div class="mt-2 space-y-2">
            <div
              v-for="(pair, index) in mappingPairs"
              :key="index"
              class="flex items-center gap-2"
            >
              <Input
                v-model="pair.key"
                :placeholder="t('editor.key')"
                class="h-7 text-xs"
                @blur="syncMappings"
              />
              <Input
                v-model="pair.value"
                :placeholder="t('editor.value')"
                class="h-7 text-xs"
                @blur="syncMappings"
              />
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 shrink-0"
                @click="removeMappingPair(index)"
              >
                <X class="h-3 w-3" />
              </Button>
            </div>
            <p
              v-if="mappingPairs.length === 0"
              class="text-xs text-muted-foreground"
            >
              No input mappings configured.
            </p>
          </div>
        </div>

        <Separator />

        <!-- Config -->
        <div>
          <Label class="text-xs text-muted-foreground">{{ t('editor.config') }}</Label>
          <textarea
            v-model="configText"
            class="mt-1 w-full rounded-md border bg-transparent px-3 py-2 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-ring"
            rows="6"
            @blur="onConfigBlur"
          />
        </div>
      </div>
    </ScrollArea>

    <!-- Delete step -->
    <div class="border-t p-4">
      <Button
        variant="destructive"
        class="w-full"
        @click="onDelete"
      >
        <Trash2 class="mr-2 h-4 w-4" />
        {{ t('editor.deleteStep') }}
      </Button>
    </div>
  </div>
</template>

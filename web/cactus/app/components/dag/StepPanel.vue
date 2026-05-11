<script setup lang="ts">
import { X, Trash2 } from 'lucide-vue-next'
import type { Node } from '@vue-flow/core'
import type { StepData } from '~/composables/useDagEditor'
import type { WorkType } from '~/composables/useWorkers'
import DynamicSettingsForm from '~/components/forms/DynamicSettingsForm.vue'
import { Button } from '~/components/ui/button'
import { Badge } from '~/components/ui/badge'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'

const props = defineProps<{
  node: Node
  workTypes: WorkType[]
  readOnly?: boolean
}>()

const emit = defineEmits<{
  close: []
  openEditor: [nodeId: string]
  updateData: [nodeId: string, data: Partial<StepData>]
  updateStep: [stepId: string, data: Record<string, unknown>]
  deleteStep: [stepId: string]
}>()

const { t } = useI18n()
const controlSteps = useControlSteps()

const isStart = computed(() => props.node.data.controlKind === 'start')
const controlDefinition = computed(() =>
  controlSteps.find(control => control.kind === props.node.data.controlKind),
)
const settingsSchema = computed(() =>
  controlDefinition.value?.settingsSchema ?? { type: 'object', properties: {} },
)
const formData = ref<Record<string, unknown>>({})

watch(
  () => props.node.id,
  () => {
    formData.value = { ...(props.node.data.controlSettings ?? {}) }
  },
  { immediate: true },
)

function onSave() {
  if (props.readOnly || isStart.value) return
  emit('updateData', props.node.id, { controlSettings: { ...formData.value } })
  emit('updateStep', props.node.id, { control_settings: { ...formData.value } })
}

function onDelete() {
  if (props.readOnly || isStart.value) return
  emit('deleteStep', props.node.id)
}
</script>

<template>
  <div class="flex w-[360px] flex-col border-l bg-background">
    <div class="flex items-center justify-between border-b px-4 py-3">
      <div class="min-w-0">
        <p class="truncate text-sm font-semibold">
          {{ node.data.label }}
        </p>
        <div class="mt-1 flex items-center gap-2">
          <Badge variant="secondary">
            {{ isStart ? 'Start' : t('editor.control') }}
          </Badge>
          <Badge v-if="node.data.controlKind" variant="outline">
            {{ node.data.controlKind }}
          </Badge>
        </div>
      </div>
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
        <div v-if="isStart" class="rounded-md border bg-muted/40 p-3 text-sm text-muted-foreground">
          {{ t('editor.startStepReadOnly') }}
        </div>

        <template v-else>
          <div>
            <p class="text-xs font-medium text-muted-foreground">
              {{ t('nodeEditor.parameters') }}
            </p>
            <div class="mt-3">
              <DynamicSettingsForm
                :schema="settingsSchema"
                :model-value="formData"
                :disabled="readOnly"
                @update:model-value="formData = $event"
              />
            </div>
          </div>
        </template>
      </div>
    </ScrollArea>

    <div v-if="!isStart" class="space-y-3 border-t p-4">
      <Button class="w-full" :disabled="readOnly" @click="onSave">
        {{ t('common.save') }}
      </Button>
      <Separator />
      <Button
        v-if="!readOnly"
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

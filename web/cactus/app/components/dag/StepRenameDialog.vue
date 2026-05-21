<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import {
  isStepNameDuplicate,
  normalizeStepName,
} from '~/components/dag/step-name-utils'
import type { StepNameEntry } from '~/components/dag/step-name-utils'

const props = defineProps<{
  open: boolean
  nodeId: string | null
  currentName: string
  names: StepNameEntry[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [nodeId: string, name: string]
}>()

const { t } = useI18n()
const draftName = ref('')

watch(
  () => [props.open, props.currentName],
  () => {
    if (props.open) {
      draftName.value = props.currentName
    }
  },
  { immediate: true },
)

const normalizedName = computed(() => normalizeStepName(draftName.value))
const isDuplicate = computed(() =>
  props.nodeId
    ? isStepNameDuplicate(normalizedName.value, props.names, props.nodeId)
    : false,
)

const errorMessage = computed(() => {
  if (!normalizedName.value) return t('editor.stepNameRequired')
  if (isDuplicate.value) return t('editor.stepNameDuplicate')
  return ''
})

const canSave = computed(() =>
  Boolean(props.nodeId && normalizedName.value && !isDuplicate.value),
)

function onSave() {
  if (!props.nodeId || !canSave.value) return
  emit('save', props.nodeId, normalizedName.value)
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ t('editor.renameStep') }}</DialogTitle>
      </DialogHeader>

      <div class="space-y-2">
        <Input
          v-model="draftName"
          data-testid="step-name-input"
          :aria-label="t('editor.stepName')"
          @keydown.enter.prevent="onSave"
        />
        <p
          v-if="errorMessage"
          data-testid="step-name-error"
          class="text-sm text-destructive"
        >
          {{ errorMessage }}
        </p>
      </div>

      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          @click="emit('update:open', false)"
        >
          {{ t('common.cancel') }}
        </Button>
        <Button
          type="button"
          data-testid="save-step-name"
          :disabled="!canSave"
          @click="onSave"
        >
          {{ t('common.save') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

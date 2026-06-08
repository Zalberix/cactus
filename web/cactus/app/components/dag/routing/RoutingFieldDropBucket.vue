<script setup lang="ts">
import type { WorkflowInputField } from '~/components/dag/node-editor/workflow-input-utils'
import { Badge } from '~/components/ui/badge'

const props = defineProps<{
  field: WorkflowInputField
  label?: string
  error?: string
}>()

const emit = defineEmits<{
  dropField: [target: WorkflowInputField]
  select: [field: WorkflowInputField]
}>()

function onDrop(event: DragEvent) {
  event.preventDefault()
  emit('dropField', props.field)
}
</script>

<template>
  <button
    type="button"
    class="w-full rounded-xl border bg-background p-3 text-left text-sm transition hover:border-primary"
    :class="{ 'border-destructive bg-destructive/5': error }"
    :data-testid="`routing-target-bucket-${field.name}`"
    @click="emit('select', field)"
    @dragover.prevent
    @drop="onDrop"
  >
    <div class="flex items-center justify-between gap-2">
      <span class="font-medium">{{ field.name }}</span>
      <Badge variant="outline">{{ field.type }}</Badge>
    </div>
    <div class="mt-1 text-xs text-muted-foreground">{{ label || '-' }}</div>
    <div v-if="error" class="mt-2 text-xs text-destructive">{{ error }}</div>
  </button>
</template>

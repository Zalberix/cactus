<script setup lang="ts">
import type { WorkflowInputField } from '~/components/dag/node-editor/workflow-input-utils'
import { Badge } from '~/components/ui/badge'

defineProps<{
  title: string
  fields: WorkflowInputField[]
  draggable?: boolean
}>()

const emit = defineEmits<{
  dragField: [field: WorkflowInputField]
  selectField: [field: WorkflowInputField]
}>()
</script>

<template>
  <div class="rounded-2xl border bg-background p-4">
    <h2 class="font-medium">{{ title }}</h2>
    <div class="mt-4 space-y-2">
      <button
        v-for="field in fields"
        :key="field.name"
        type="button"
        class="flex w-full items-center justify-between rounded-xl border bg-muted/30 px-3 py-2 text-left text-sm"
        :draggable="draggable"
        :data-testid="`routing-field-${field.name}`"
        @dragstart="emit('dragField', field)"
        @click="emit('selectField', field)"
      >
        <span>
          <span class="font-medium">{{ field.name }}</span>
          <span v-if="field.description" class="ml-2 text-xs text-muted-foreground">{{ field.description }}</span>
        </span>
        <Badge variant="outline">{{ field.type }}</Badge>
      </button>
    </div>
  </div>
</template>

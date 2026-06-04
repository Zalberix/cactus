<script setup lang="ts">
import type {
  WorkflowExperiment,
  WorkflowInputSchemaRecord,
} from '~/composables/useWorkflowRouting'
import type { RoutingScopeForm } from './types'
import { computed } from 'vue'
import { Button } from '~/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import { Input } from '~/components/ui/input'

const props = defineProps<{
  open: boolean
  form: RoutingScopeForm
  experiments: WorkflowExperiment[]
  activeSchemas: WorkflowInputSchemaRecord[]
  saving?: boolean
  editingId: number | null
  schemaLabel: (schemaId: number) => string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: []
  cancel: []
}>()

const { t } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

function cancel() {
  emit('update:open', false)
  emit('cancel')
}
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-3xl">
      <div data-testid="scope-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.updateScope') : t('workflowRouting.createScope') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.experimentsDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperiment') }}</span>
            <select v-model="form.experimentId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">Select experiment</option>
              <option v-for="experiment in experiments" :key="experiment.id" :value="String(experiment.id)">{{ experiment.name }}</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldInputSchema') }}</span>
            <select v-model="form.inputSchemaId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">Select input schema</option>
              <option v-for="schema in activeSchemas" :key="schema.id" :value="String(schema.id)">{{ schemaLabel(schema.id) }}</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldTrafficPercent') }}</span>
            <Input v-model.number="form.trafficPercent" type="number" min="0" max="100" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldTrafficConditionsJson') }}</span>
            <textarea v-model="form.trafficConditionsJson" class="min-h-24 w-full rounded-md border bg-background p-3 font-mono text-xs" />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving || !form.experimentId || !form.inputSchemaId">
              {{ editingId ? t('workflowRouting.updateScope') : t('workflowRouting.createScope') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

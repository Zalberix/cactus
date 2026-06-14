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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

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
const EMPTY_SELECT_VALUE = '__empty__'

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const experimentSelectValue = computed({
  get: () => props.form.experimentId || EMPTY_SELECT_VALUE,
  set: (value: string) => {
    props.form.experimentId = value === EMPTY_SELECT_VALUE ? '' : value
  },
})

const inputSchemaSelectValue = computed({
  get: () => props.form.inputSchemaId || EMPTY_SELECT_VALUE,
  set: (value: string) => {
    props.form.inputSchemaId = value === EMPTY_SELECT_VALUE ? '' : value
  },
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
          <DialogDescription>{{ t('workflowRouting.testingDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperiment') }}</span>
            <Select v-model="experimentSelectValue">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="EMPTY_SELECT_VALUE">Select experiment</SelectItem>
                <SelectItem v-for="experiment in experiments" :key="experiment.id" :value="String(experiment.id)">
                  {{ experiment.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldInputSchema') }}</span>
            <Select v-model="inputSchemaSelectValue">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="EMPTY_SELECT_VALUE">Select input schema</SelectItem>
                <SelectItem v-for="schema in activeSchemas" :key="schema.id" :value="String(schema.id)">
                  {{ schemaLabel(schema.id) }}
                </SelectItem>
              </SelectContent>
            </Select>
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

<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type {
  WorkflowExperiment,
  WorkflowExperimentScope,
} from '~/composables/useWorkflowRouting'
import type { RoutingVariantForm } from './types'
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
  form: RoutingVariantForm
  experiments: WorkflowExperiment[]
  scopesByExperiment: Record<number, WorkflowExperimentScope[]>
  activeVersions: VersionSummary[]
  saving?: boolean
  editingId: number | null
  versionLabel: (versionId: number) => string
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
      <div data-testid="variant-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.updateVariant') : t('workflowRouting.createVariant') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.testingDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldScope') }}</span>
            <select v-model="form.scopeId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">Select scope</option>
              <option v-for="experiment in experiments" :key="`exp-${experiment.id}`" disabled>{{ experiment.name }}</option>
              <template v-for="experiment in experiments" :key="experiment.id">
                <option v-for="scope in scopesByExperiment[experiment.id] || []" :key="scope.id" :value="String(scope.id)">
                  {{ schemaLabel(scope.workflow_input_schema_id) }} &middot; {{ t('workflowRouting.scopeNumber', { id: scope.id }) }}
                </option>
              </template>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldWorkflowVersion') }}</span>
            <select v-model="form.versionId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">Select version</option>
              <option v-for="version in activeVersions" :key="version.id" :value="String(version.id)">{{ versionLabel(version.id) }}</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldTrafficWeight') }}</span>
            <Input v-model.number="form.trafficWeight" type="number" min="0" max="100" />
          </label>
          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block space-y-1.5">
              <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldControlGroup') }}</span>
              <span class="flex items-center gap-2 text-sm">
                <input v-model="form.isControlGroup" type="checkbox">
                {{ t('workflowRouting.fieldEnabled') }}
              </span>
            </label>
            <label class="block space-y-1.5">
              <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldActiveVariant') }}</span>
              <span class="flex items-center gap-2 text-sm">
                <input v-model="form.isActive" type="checkbox">
                {{ t('workflowRouting.fieldEnabled') }}
              </span>
            </label>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving || !form.scopeId || !form.versionId">
              {{ editingId ? t('workflowRouting.updateVariant') : t('workflowRouting.createVariant') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

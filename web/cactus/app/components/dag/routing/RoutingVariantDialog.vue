<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type {
  WorkflowExperiment,
  WorkflowExperimentScope,
} from '~/composables/useWorkflowRouting'
import type { RoutingVariantForm } from './types'
import { computed } from 'vue'
import { Button } from '~/components/ui/button'
import { Checkbox } from '~/components/ui/checkbox'
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
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

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
const EMPTY_SELECT_VALUE = '__empty__'

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const scopeSelectValue = computed({
  get: () => props.form.scopeId || EMPTY_SELECT_VALUE,
  set: (value: string) => {
    props.form.scopeId = value === EMPTY_SELECT_VALUE ? '' : value
  },
})

const versionSelectValue = computed({
  get: () => props.form.versionId || EMPTY_SELECT_VALUE,
  set: (value: string) => {
    props.form.versionId = value === EMPTY_SELECT_VALUE ? '' : value
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
      <div data-testid="variant-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.updateVariant') : t('workflowRouting.createVariant') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.testingDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldScope') }}</span>
            <Select v-model="scopeSelectValue">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="EMPTY_SELECT_VALUE">Select scope</SelectItem>
                <SelectGroup v-for="experiment in experiments" :key="experiment.id">
                  <SelectLabel>{{ experiment.name }}</SelectLabel>
                  <SelectItem
                    v-for="scope in scopesByExperiment[experiment.id] || []"
                    :key="scope.id"
                    :value="String(scope.id)"
                  >
                    {{ schemaLabel(scope.workflow_input_schema_id) }} &middot; {{ t('workflowRouting.scopeNumber', { id: scope.id }) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldWorkflowVersion') }}</span>
            <Select v-model="versionSelectValue">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="EMPTY_SELECT_VALUE">Select version</SelectItem>
                <SelectItem v-for="version in activeVersions" :key="version.id" :value="String(version.id)">
                  {{ versionLabel(version.id) }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldTrafficWeight') }}</span>
            <Input v-model.number="form.trafficWeight" type="number" min="0" max="100" />
          </label>
          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block space-y-1.5">
              <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldControlGroup') }}</span>
              <span class="flex items-center gap-2 text-sm">
                <Checkbox v-model="form.isControlGroup" />
                {{ t('workflowRouting.fieldEnabled') }}
              </span>
            </label>
            <label class="block space-y-1.5">
              <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldActiveVariant') }}</span>
              <span class="flex items-center gap-2 text-sm">
                <Checkbox v-model="form.isActive" />
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

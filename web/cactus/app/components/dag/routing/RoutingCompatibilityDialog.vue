<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type {
  WorkflowInputMapperRecord,
  WorkflowInputSchemaRecord,
} from '~/composables/useWorkflowRouting'
import type { RoutingCompatibilityForm } from './types'
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
  form: RoutingCompatibilityForm
  activeVersions: VersionSummary[]
  activeSchemas: WorkflowInputSchemaRecord[]
  mappers: WorkflowInputMapperRecord[]
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
      <div data-testid="compatibility-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.editCompatibility') : t('workflowRouting.createCompatibility') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.compatibilitiesDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldWorkflowVersion') }}</span>
            <select v-model="form.versionId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">Select version</option>
              <option v-for="version in activeVersions" :key="version.id" :value="String(version.id)">{{ versionLabel(version.id) }}</option>
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
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldInputMapper') }}</span>
            <select v-model="form.mapperId" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="">No mapper</option>
              <option v-for="mapper in mappers" :key="mapper.id" :value="String(mapper.id)">Mapper #{{ mapper.id }}</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldCompatibilityType') }}</span>
            <Input v-model="form.compatibilityType" :placeholder="t('workflowRouting.placeholderCompatibilityType')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldDefaultRoute') }}</span>
            <span class="flex items-center gap-2 text-sm">
              <input v-model="form.isDefaultRoute" type="checkbox">
              {{ t('workflowRouting.fieldEnabled') }}
            </span>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldDefaultValuesJson') }}</span>
            <textarea v-model="form.defaultValuesJson" class="min-h-24 w-full rounded-md border bg-background p-3 font-mono text-xs" />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving || !form.versionId || !form.inputSchemaId">
              {{ editingId ? t('workflowRouting.updateCompatibility') : t('workflowRouting.createCompatibility') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

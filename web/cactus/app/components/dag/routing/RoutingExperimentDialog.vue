<script setup lang="ts">
import type { RoutingExperimentForm } from './types'
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
  form: RoutingExperimentForm
  saving?: boolean
  editingId: number | null
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
    <DialogContent class="max-w-2xl">
      <div data-testid="experiment-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.editExperiment') : t('workflowRouting.createExperiment') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.experimentsDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperimentName') }}</span>
            <Input v-model="form.name" :placeholder="t('workflowRouting.placeholderExperimentName')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperimentType') }}</span>
            <select v-model="form.experimentType" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="split">{{ t('workflowRouting.typeSplit') }}</option>
              <option value="canary">{{ t('workflowRouting.typeCanary') }}</option>
              <option value="rollout">{{ t('workflowRouting.typeRollout') }}</option>
              <option value="shadow">{{ t('workflowRouting.typeShadow') }}</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldStatus') }}</span>
            <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="draft">{{ t('workflowRouting.statusDraft') }}</option>
              <option value="active">{{ t('workflowRouting.statusActive') }}</option>
              <option value="paused">{{ t('workflowRouting.statusPaused') }}</option>
            </select>
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving">
              {{ editingId ? t('workflowRouting.updateExperiment') : t('workflowRouting.createExperiment') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

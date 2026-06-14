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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

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
          <DialogTitle>{{ editingId ? t('workflowRouting.editTesting') : t('workflowRouting.createTesting') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.testingDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperimentName') }}</span>
            <Input v-model="form.name" :placeholder="t('workflowRouting.placeholderExperimentName')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldExperimentType') }}</span>
            <Select v-model="form.experimentType">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="split">{{ t('workflowRouting.typeSplit') }}</SelectItem>
                <SelectItem value="canary">{{ t('workflowRouting.typeCanary') }}</SelectItem>
                <SelectItem value="rollout">{{ t('workflowRouting.typeRollout') }}</SelectItem>
                <SelectItem value="shadow">{{ t('workflowRouting.typeShadow') }}</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldStatus') }}</span>
            <Select v-model="form.status">
              <SelectTrigger class="h-10 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="draft">{{ t('workflowRouting.statusDraft') }}</SelectItem>
                <SelectItem value="active">{{ t('workflowRouting.statusActive') }}</SelectItem>
                <SelectItem value="paused">{{ t('workflowRouting.statusPaused') }}</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving">
              {{ editingId ? t('workflowRouting.updateTesting') : t('workflowRouting.createTesting') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

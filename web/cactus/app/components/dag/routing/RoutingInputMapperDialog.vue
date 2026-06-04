<script setup lang="ts">
import type { RoutingMapperForm } from './types'
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
  form: RoutingMapperForm
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
    <DialogContent class="max-w-3xl">
      <div data-testid="mapper-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.editMapper') : t('workflowRouting.createMapper') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.inputMappersDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldMapperName') }}</span>
            <Input v-model="form.name" :placeholder="t('workflowRouting.placeholderMapperName')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldMapperType') }}</span>
            <Input v-model="form.mapperType" :placeholder="t('workflowRouting.placeholderMapperType')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldMapperRulesJson') }}</span>
            <textarea v-model="form.rulesJson" class="min-h-32 w-full rounded-md border bg-background p-3 font-mono text-xs" />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving">
              {{ editingId ? t('workflowRouting.updateMapper') : t('workflowRouting.createMapper') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

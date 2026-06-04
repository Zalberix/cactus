<script setup lang="ts">
import type { RoutingDeleteTarget } from './types'
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

const props = defineProps<{
  open: boolean
  target: RoutingDeleteTarget | null
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  cancel: []
  confirm: []
}>()

const { t } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-md">
      <div data-testid="routing-delete-dialog">
        <DialogHeader>
          <DialogTitle>{{ t('common.delete') }}</DialogTitle>
          <DialogDescription>{{ target?.message }}</DialogDescription>
        </DialogHeader>
        <DialogFooter class="mt-4">
          <Button type="button" variant="outline" :disabled="saving" @click="emit('cancel')">
            {{ t('destructive.cancel') }}
          </Button>
          <Button type="button" variant="destructive" :disabled="saving || !target" data-testid="confirm-routing-delete" @click="emit('confirm')">
            {{ saving ? t('common.loading') : t('destructive.delete') }}
          </Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>

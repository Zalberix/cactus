<script setup lang="ts">
import type { RoutingVersionRowRecord } from '~/composables/useWorkflowRouting'
import { computed, ref, watch } from 'vue'
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
  row: RoutingVersionRowRecord | null
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [requiredLabel: string]
}>()

const { t } = useI18n()
const confirmation = ref('')

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const requiredLabel = computed(() => props.row
  ? `${props.row.input_schema_code} v${props.row.input_schema_version_number}`
  : '')

watch(() => props.open, (open) => {
  if (!open) confirmation.value = ''
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent>
      <div data-testid="routing-archive-schema-dialog">
        <DialogHeader>
          <DialogTitle>{{ t('workflowRouting.archiveSchema') }}</DialogTitle>
          <DialogDescription>
            {{ t('workflowRouting.archiveSchemaConfirm', { name: requiredLabel }) }}
          </DialogDescription>
        </DialogHeader>

        <label class="mt-4 block space-y-2">
          <span class="text-sm font-medium">{{ t('workflowRouting.archiveSchemaTypeName') }}</span>
          <Input
            v-model="confirmation"
            data-testid="routing-archive-confirmation"
            :placeholder="requiredLabel"
            :disabled="saving"
          />
        </label>

        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" :disabled="saving" @click="isOpen = false">
            {{ t('common.cancel') }}
          </Button>
          <Button
            type="button"
            variant="destructive"
            data-testid="routing-archive-confirm"
            :disabled="saving || confirmation !== requiredLabel"
            @click="emit('confirm', requiredLabel)"
          >
            {{ t('workflowRouting.archiveSchema') }}
          </Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>

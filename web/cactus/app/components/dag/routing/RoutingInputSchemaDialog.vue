<script setup lang="ts">
import type { RoutingSchemaForm } from './types'
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
  form: RoutingSchemaForm
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
      <div data-testid="schema-routing-dialog">
        <DialogHeader>
          <DialogTitle>{{ editingId ? t('workflowRouting.editInputSchema') : t('workflowRouting.createInputSchema') }}</DialogTitle>
          <DialogDescription>{{ t('workflowRouting.inputSchemasDescription') }}</DialogDescription>
        </DialogHeader>
        <form class="mt-4 space-y-3" @submit.prevent="emit('submit')">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldSchemaCode') }}</span>
            <Input v-model="form.code" data-testid="schema-code-input" :placeholder="t('workflowRouting.placeholderSchemaCode')" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldStatus') }}</span>
            <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
              <option value="draft">{{ t('workflowRouting.statusDraft') }}</option>
              <option value="active">{{ t('workflowRouting.statusActive') }}</option>
              <option value="deprecated">deprecated</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldDefaultSchema') }}</span>
            <span class="flex items-center gap-2 text-sm">
              <input v-model="form.isDefault" type="checkbox">
              {{ t('workflowRouting.fieldEnabled') }}
            </span>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldSchemaJson') }}</span>
            <textarea v-model="form.schemaJson" class="min-h-40 w-full rounded-md border bg-background p-3 font-mono text-xs" />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="cancel">
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit" :disabled="saving">
              {{ editingId ? t('workflowRouting.updateInputSchema') : t('workflowRouting.createInputSchema') }}
            </Button>
          </DialogFooter>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>

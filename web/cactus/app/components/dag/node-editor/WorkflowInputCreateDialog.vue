<script setup lang="ts">
import type { WorkflowInputField } from '~/composables/useVersions'
import { Button } from '~/components/ui/button'
import { Checkbox } from '~/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  create: [data: Omit<WorkflowInputField, 'id' | 'workflow_version_id'>]
}>()

const { t } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const name = ref('')
const type = ref<WorkflowInputField['type']>('string')
const required = ref(true)
const description = ref('')

function onCreate() {
  if (!name.value.trim()) return
  emit('create', {
    name: name.value.trim(),
    type: type.value,
    required: required.value,
    description: description.value.trim() || undefined,
  })
  name.value = ''
  type.value = 'string'
  required.value = true
  description.value = ''
}
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ t('nodeEditor.newWorkflowInput') }}</DialogTitle>
      </DialogHeader>

      <div class="space-y-4">
        <div class="space-y-1">
          <Label>{{ t('workflows.name') }}</Label>
          <Input v-model="name" />
        </div>

        <div class="space-y-1">
          <Label>{{ t('editor.stepType') }}</Label>
          <select v-model="type" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
            <option value="string">string</option>
            <option value="number">number</option>
            <option value="integer">integer</option>
            <option value="boolean">boolean</option>
            <option value="object">object</option>
            <option value="array">array</option>
          </select>
        </div>

        <label class="flex items-center gap-2 text-sm">
          <Checkbox v-model:checked="required" />
          {{ t('nodeEditor.required') }}
        </label>

        <div class="space-y-1">
          <Label>{{ t('roles.description') }}</Label>
          <Input v-model="description" />
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false">{{ t('common.cancel') }}</Button>
        <Button :disabled="!name.trim()" @click="onCreate">{{ t('common.create') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import type { WorkflowInputSchemaField } from '~/composables/useWorkflows'
import { Badge } from '~/components/ui/badge'
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
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import type { WorkflowInputField } from './node-editor/workflow-input-utils'
import { workflowInputFieldsFromSchema } from './node-editor/workflow-input-utils'

const props = defineProps<{
  schema: Record<string, unknown> | null
  open: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saveField: [field: WorkflowInputSchemaField]
  deleteField: [field: WorkflowInputField]
}>()

const { t } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const inputTypes: WorkflowInputSchemaField['type'][] = ['string', 'number', 'integer', 'boolean', 'object', 'array']
const fields = computed(() => workflowInputFieldsFromSchema(props.schema))
const formOpen = ref(false)
const editingName = ref<string | null>(null)
const deleteTarget = ref<WorkflowInputField | null>(null)
const form = ref<WorkflowInputSchemaField>({
  name: '',
  type: 'string',
  required: false,
  description: '',
})

const isEditing = computed(() => editingName.value !== null)
const canSubmit = computed(() => form.value.name.trim().length > 0 && !props.saving)

function startCreate() {
  deleteTarget.value = null
  editingName.value = null
  form.value = {
    name: '',
    type: 'string',
    required: false,
    description: '',
  }
  formOpen.value = true
}

function startEdit(field: WorkflowInputField) {
  deleteTarget.value = null
  editingName.value = field.name
  form.value = {
    name: field.name,
    type: field.type,
    required: field.required,
    description: field.description ?? '',
  }
  formOpen.value = true
}

function submitField() {
  if (!canSubmit.value) return
  const description = form.value.description?.trim()
  emit('saveField', {
    name: form.value.name.trim(),
    type: form.value.type,
    required: form.value.required,
    description: description || undefined,
  })
}

function confirmDelete(field: WorkflowInputField) {
  formOpen.value = false
  deleteTarget.value = field
}

function deleteField() {
  if (!deleteTarget.value || props.saving) return
  emit('deleteField', deleteTarget.value)
}

watch(() => props.schema, () => {
  formOpen.value = false
  deleteTarget.value = null
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-4xl">
      <DialogHeader>
        <DialogTitle>{{ t('editor.inputSchema') }}</DialogTitle>
        <DialogDescription>{{ t('editor.inputSchemaDescription') }}</DialogDescription>
      </DialogHeader>

      <div class="flex items-center justify-end">
        <Button type="button" :disabled="saving" @click="startCreate">
          <Plus class="h-4 w-4" />
          {{ t('workflowInputs.addInput') }}
        </Button>
      </div>

      <div class="max-h-[420px] overflow-auto rounded-md border">
        <Table data-testid="workflow-input-schema-table">
          <TableHeader>
            <TableRow>
              <TableHead>{{ t('workflowInputs.name') }}</TableHead>
              <TableHead>{{ t('workflowInputs.type') }}</TableHead>
              <TableHead>{{ t('workflowInputs.required') }}</TableHead>
              <TableHead>{{ t('workflowInputs.description') }}</TableHead>
              <TableHead class="w-24 text-right">{{ t('workflowInputs.actions') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="fields.length === 0" :colspan="5">
              {{ t('workflowInputs.empty') }}
            </TableEmpty>
            <TableRow
              v-for="field in fields"
              :key="field.name"
              :data-testid="`workflow-input-row-${field.name}`"
            >
              <TableCell class="font-medium">{{ field.name }}</TableCell>
              <TableCell :data-testid="`workflow-input-type-${field.name}`">
                <Badge variant="secondary">{{ field.type }}</Badge>
              </TableCell>
              <TableCell :data-testid="`workflow-input-required-${field.name}`">
                <Badge v-if="field.required" variant="outline">{{ t('workflowInputs.required') }}</Badge>
                <span v-else class="text-sm text-muted-foreground">{{ t('workflowInputs.optional') }}</span>
              </TableCell>
              <TableCell class="max-w-[280px] truncate text-sm text-muted-foreground">
                {{ field.description || t('workflowInputs.noDescription') }}
              </TableCell>
              <TableCell>
                <div class="flex justify-end gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8"
                    :aria-label="t('workflowInputs.editInput')"
                    :title="t('workflowInputs.editInput')"
                    :disabled="saving"
                    @click="startEdit(field)"
                  >
                    <Pencil class="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 text-destructive hover:text-destructive"
                    :aria-label="t('workflowInputs.deleteInput')"
                    :title="t('workflowInputs.deleteInput')"
                    :disabled="saving"
                    @click="confirmDelete(field)"
                  >
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <form v-if="formOpen" class="rounded-md border bg-muted/30 p-4" @submit.prevent="submitField">
        <div class="grid gap-3 md:grid-cols-[1.2fr_0.8fr_auto] md:items-end">
          <label class="space-y-1">
            <span class="text-xs font-medium">{{ t('workflowInputs.name') }}</span>
            <Input
              v-model="form.name"
              :readonly="isEditing"
              :disabled="saving"
            />
          </label>
          <label class="space-y-1">
            <span class="text-xs font-medium">{{ t('workflowInputs.type') }}</span>
            <Select v-model="form.type" :disabled="saving">
              <SelectTrigger class="h-9 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="type in inputTypes" :key="type" :value="type">
                  {{ type }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>
          <label class="flex h-9 items-center gap-2 text-sm">
            <Checkbox v-model="form.required" :disabled="saving" />
            <span>{{ t('workflowInputs.required') }}</span>
          </label>
        </div>
        <label class="mt-3 block space-y-1">
          <span class="text-xs font-medium">{{ t('workflowInputs.description') }}</span>
          <Input v-model="form.description" class="h-9" :disabled="saving" />
        </label>
        <DialogFooter class="mt-4">
          <Button type="button" variant="outline" :disabled="saving" @click="formOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button type="submit" :disabled="!canSubmit">
            {{ saving ? t('common.loading') : t('common.save') }}
          </Button>
        </DialogFooter>
      </form>

      <div v-if="deleteTarget" class="rounded-md border border-destructive/30 bg-destructive/5 p-4">
        <p class="text-sm font-medium">
          {{ t('workflowInputs.deleteConfirmTitle', { name: deleteTarget.name }) }}
        </p>
        <p class="mt-1 text-sm text-muted-foreground">
          {{ t('workflowInputs.deleteConfirmBody') }}
        </p>
        <DialogFooter class="mt-4">
          <Button type="button" variant="outline" :disabled="saving" @click="deleteTarget = null">
            {{ t('destructive.cancel') }}
          </Button>
          <Button type="button" variant="destructive" :disabled="saving" @click="deleteField">
            {{ saving ? t('common.loading') : t('destructive.delete') }}
          </Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>

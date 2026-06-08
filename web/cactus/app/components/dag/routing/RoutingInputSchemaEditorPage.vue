<script setup lang="ts">
import type { WorkflowInputSchemaField } from '~/composables/useWorkflows'
import type { WorkflowInputSchemaRecord } from '~/composables/useWorkflowRouting'
import type { WorkflowInputField } from '~/components/dag/node-editor/workflow-input-utils'
import { ArrowLeft, FileJson, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import {
  deleteWorkflowInputFieldFromSchema,
  upsertWorkflowInputFieldInSchema,
  workflowInputFieldsFromSchema,
} from '~/components/dag/node-editor/workflow-input-utils'
import {
  workflowRoutingPath,
} from '~/composables/useWorkflowRouting'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import { toast } from '~/components/ui/toast/use-toast'

const props = defineProps<{
  mode: 'create' | 'edit'
}>()

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const routing = useWorkflowRouting()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))
const inputSchemaId = computed(() => Number(route.params.inputSchemaId || 0))
const isEdit = computed(() => props.mode === 'edit')

const inputTypes: WorkflowInputSchemaField['type'][] = ['string', 'number', 'integer', 'boolean', 'object', 'array']
const loading = ref(isEdit.value)
const saving = ref(false)
const schema = ref<Record<string, unknown>>(emptySchema())
const loadedSchema = ref<WorkflowInputSchemaRecord | null>(null)
const fields = computed(() => workflowInputFieldsFromSchema(schema.value))

const form = reactive({
  code: 'public',
  status: 'active',
  isDefault: false,
})

const fieldFormOpen = ref(false)
const editingFieldName = ref<string | null>(null)
const deletingField = ref<WorkflowInputField | null>(null)
const fieldForm = ref<WorkflowInputSchemaField>({
  name: '',
  type: 'string',
  required: false,
  description: '',
})

const isReadonly = computed(() => isEdit.value && loadedSchema.value?.usage?.is_readonly === true)
const isLocked = computed(() => saving.value || isReadonly.value)
const canSaveField = computed(() => fieldForm.value.name.trim().length > 0 && !isLocked.value)
const canSaveSchema = computed(() => form.code.trim().length > 0 && !loading.value && !isLocked.value)
const pageTitle = computed(() => isEdit.value ? t('workflowRouting.editInputSchema') : t('workflowRouting.createInputSchema'))
const saveLabel = computed(() => isEdit.value ? t('workflowRouting.updateInputSchema') : t('workflowRouting.createInputSchema'))
const readonlyReasons = computed(() => loadedSchema.value?.usage?.reasons ?? [])

function emptySchema(): Record<string, unknown> {
  return {
    type: 'object',
    properties: {},
  }
}

function routingListPath() {
  return workflowRoutingPath(orgId.value, workflowId.value)
}

async function loadInputSchema() {
  if (!isEdit.value || !inputSchemaId.value) return
  loading.value = true
  try {
    const record = await routing.fetchInputSchema(inputSchemaId.value)
    loadedSchema.value = record
    form.code = record.code
    form.status = record.status
    form.isDefault = record.is_default
    schema.value = record.schema_json ?? emptySchema()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function goBack() {
  router.push(routingListPath())
}

function startCreateField() {
  if (isReadonly.value) return
  deletingField.value = null
  editingFieldName.value = null
  fieldForm.value = {
    name: '',
    type: 'string',
    required: false,
    description: '',
  }
  fieldFormOpen.value = true
}

function startEditField(field: WorkflowInputField) {
  if (isReadonly.value) return
  deletingField.value = null
  editingFieldName.value = field.name
  fieldForm.value = {
    name: field.name,
    type: field.type,
    required: field.required,
    description: field.description ?? '',
  }
  fieldFormOpen.value = true
}

function saveField() {
  if (isReadonly.value || !canSaveField.value) return
  const fieldName = fieldForm.value.name.trim()
  const description = fieldForm.value.description?.trim()
  const baseSchema = editingFieldName.value && editingFieldName.value !== fieldName
    ? deleteWorkflowInputFieldFromSchema(schema.value, editingFieldName.value)
    : schema.value
  schema.value = upsertWorkflowInputFieldInSchema(baseSchema, {
    name: fieldName,
    type: fieldForm.value.type,
    required: fieldForm.value.required,
    description: description || undefined,
  })
  fieldFormOpen.value = false
  editingFieldName.value = null
}

function confirmDeleteField(field: WorkflowInputField) {
  if (isReadonly.value) return
  fieldFormOpen.value = false
  deletingField.value = field
}

function deleteField() {
  if (isReadonly.value || !deletingField.value || saving.value) return
  schema.value = deleteWorkflowInputFieldFromSchema(schema.value, deletingField.value.name)
  deletingField.value = null
}

function readonlyReasonLabel(reason: string) {
  switch (reason) {
    case 'message':
      return t('workflowRouting.readonlyReasonMessage')
    case 'mapper':
      return t('workflowRouting.readonlyReasonMapper')
    case 'experiment':
      return t('workflowRouting.readonlyReasonTesting')
    default:
      return reason
  }
}

async function saveSchema() {
  if (isReadonly.value || !canSaveSchema.value) return
  saving.value = true
  try {
    if (isEdit.value) {
      await routing.updateInputSchema(inputSchemaId.value, {
        code: form.code.trim(),
        is_default: form.isDefault,
        schema_json: schema.value,
      })
      await routing.updateInputSchemaStatus(inputSchemaId.value, form.status)
      if (form.isDefault) {
        await routing.setInputSchemaDefault(inputSchemaId.value)
      }
      toast({ title: t('workflowRouting.updatedInputSchema') })
    }
    else {
      await routing.createInputSchema(workflowId.value, {
        code: form.code.trim(),
        status: form.status,
        is_default: form.isDefault,
        schema_json: schema.value,
      })
      toast({ title: t('workflowRouting.createdInputSchema') })
    }
    goBack()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateInputSchema')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadInputSchema()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <Button variant="ghost" class="-ml-3 mb-3 gap-2" @click="goBack">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <div class="flex items-center gap-3">
          <div class="rounded-md bg-primary/10 p-3 text-primary">
            <FileJson class="h-6 w-6" />
          </div>
          <div>
            <h1 class="text-2xl font-semibold">{{ pageTitle }}</h1>
            <p class="text-sm text-muted-foreground">{{ t('workflowRouting.inputSchemasDescription') }}</p>
          </div>
        </div>
      </div>
      <Button
        type="button"
        class="gap-2"
        :disabled="!canSaveSchema"
        data-testid="routing-schema-save"
        @click="saveSchema"
      >
        {{ saving ? t('common.loading') : saveLabel }}
      </Button>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>

    <div v-else class="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
      <div
        v-if="isReadonly"
        class="xl:col-span-2 rounded-2xl border border-amber-300/60 bg-amber-50 p-4 text-sm text-amber-950"
        data-testid="routing-schema-readonly-banner"
      >
        <div class="font-medium">{{ t('workflowRouting.readonlySchema') }}</div>
        <div class="mt-2 flex flex-wrap gap-2">
          <Badge v-for="reason in readonlyReasons" :key="reason" variant="outline">
            {{ readonlyReasonLabel(reason) }}
          </Badge>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{{ t('workflowRouting.inputSchemas') }}</CardTitle>
          <CardDescription>{{ t('workflowRouting.inputSchemasDescription') }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldSchemaCode') }}</span>
            <Input
              v-model="form.code"
              data-testid="routing-schema-code-input"
              :placeholder="t('workflowRouting.placeholderSchemaCode')"
              :disabled="isLocked"
            />
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldStatus') }}</span>
            <select
              v-model="form.status"
              class="h-10 w-full rounded-md border bg-background px-3 text-sm"
              :disabled="isLocked"
            >
              <option value="draft">{{ t('workflowRouting.statusDraft') }}</option>
              <option value="active">{{ t('workflowRouting.statusActive') }}</option>
              <option value="deprecated">deprecated</option>
            </select>
          </label>
          <label class="block space-y-1.5">
            <span class="text-sm font-medium leading-none">{{ t('workflowRouting.fieldDefaultSchema') }}</span>
            <span class="flex items-center gap-2 text-sm">
              <input v-model="form.isDefault" type="checkbox" :disabled="isLocked">
              {{ t('workflowRouting.fieldEnabled') }}
            </span>
          </label>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="space-y-3">
          <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
            <div>
              <CardTitle>{{ t('editor.inputSchema') }}</CardTitle>
              <CardDescription>{{ t('editor.inputSchemaDescription') }}</CardDescription>
            </div>
            <Button
              type="button"
              size="sm"
              class="gap-2"
              :disabled="isLocked"
              data-testid="routing-schema-add-field"
              @click="startCreateField"
            >
              <Plus class="h-4 w-4" />
              {{ t('workflowInputs.addInput') }}
            </Button>
          </div>
        </CardHeader>
        <CardContent class="space-y-4">
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
                        :disabled="isLocked"
                        :data-testid="`routing-schema-edit-field-${field.name}`"
                        @click="startEditField(field)"
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
                        :disabled="isLocked"
                        :data-testid="`routing-schema-delete-field-${field.name}`"
                        @click="confirmDeleteField(field)"
                      >
                        <Trash2 class="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <form v-if="fieldFormOpen" class="rounded-md border bg-muted/30 p-4" @submit.prevent="saveField">
            <div class="grid gap-3 md:grid-cols-[1.2fr_0.8fr_auto] md:items-end">
              <label class="space-y-1">
                <span class="text-xs font-medium">{{ t('workflowInputs.name') }}</span>
                <Input
                  v-model="fieldForm.name"
                  data-testid="routing-schema-field-name-input"
                  :disabled="isLocked"
                  class="h-9"
                />
              </label>
              <label class="space-y-1">
                <span class="text-xs font-medium">{{ t('workflowInputs.type') }}</span>
                <select
                  v-model="fieldForm.type"
                  data-testid="routing-schema-field-type-select"
                  class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  :disabled="isLocked"
                >
                  <option v-for="type in inputTypes" :key="type" :value="type">
                    {{ type }}
                  </option>
                </select>
              </label>
              <label class="flex h-9 items-center gap-2 text-sm">
                <input
                  v-model="fieldForm.required"
                  data-testid="routing-schema-field-required-checkbox"
                  type="checkbox"
                  class="h-4 w-4 rounded border-input"
                  :disabled="isLocked"
                >
                <span>{{ t('workflowInputs.required') }}</span>
              </label>
            </div>
            <label class="mt-3 block space-y-1">
              <span class="text-xs font-medium">{{ t('workflowInputs.description') }}</span>
              <Input
                v-model="fieldForm.description"
                data-testid="routing-schema-field-description-input"
                class="h-9"
                :disabled="isLocked"
              />
            </label>
            <div class="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" :disabled="isLocked" @click="fieldFormOpen = false">
                {{ t('common.cancel') }}
              </Button>
              <Button type="button" :disabled="!canSaveField" data-testid="routing-schema-save-field" @click="saveField">
                {{ t('common.save') }}
              </Button>
            </div>
          </form>

          <div v-if="deletingField" class="rounded-md border border-destructive/30 bg-destructive/5 p-4">
            <p class="text-sm font-medium">
              {{ t('workflowInputs.deleteConfirmTitle', { name: deletingField.name }) }}
            </p>
            <p class="mt-1 text-sm text-muted-foreground">
              {{ t('workflowInputs.deleteConfirmBody') }}
            </p>
            <div class="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" :disabled="isLocked" @click="deletingField = null">
                {{ t('common.cancel') }}
              </Button>
              <Button type="button" variant="destructive" :disabled="isLocked" @click="deleteField">
                {{ t('common.delete') }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

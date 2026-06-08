<script setup lang="ts">
import type { CompatibilityMappingFieldPayload, WorkflowInputSchemaRecord } from '~/composables/useWorkflowRouting'
import type { WorkflowInputField } from '~/components/dag/node-editor/workflow-input-utils'
import { ArrowLeft, Wand2, X } from 'lucide-vue-next'
import { workflowInputFieldsFromSchema } from '~/components/dag/node-editor/workflow-input-utils'
import {
  workflowInputSchemaCompatibilitiesPath,
} from '~/composables/useWorkflowRouting'
import RoutingSchemaFieldsColumn from '~/components/dag/routing/RoutingSchemaFieldsColumn.vue'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'
import { toast } from '~/components/ui/toast/use-toast'

type FieldAction =
  | { mode: 'source', sourcePath: string }
  | { mode: 'default', value: unknown }

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const routing = useWorkflowRouting()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))
const inputSchemaId = computed(() => Number(route.params.inputSchemaId))

const sourceSchema = ref<WorkflowInputSchemaRecord | null>(null)
const targetSchema = ref<WorkflowInputSchemaRecord | null>(null)
const targetResults = ref<WorkflowInputSchemaRecord[]>([])
const targetQuery = ref('')
const defaultDrafts = ref<Record<string, string>>({})
const selectedTargetField = ref<WorkflowInputField | null>(null)
const draggedSourceField = ref<WorkflowInputField | null>(null)
const dragOverTargetName = ref<string | null>(null)
const fieldActions = ref<Record<string, FieldAction>>({})
const validationErrors = ref<Record<string, string>>({})
const loading = ref(true)
const saving = ref(false)

const sourceFields = computed(() => sourceSchema.value ? workflowInputFieldsFromSchema(sourceSchema.value.schema_json) : [])
const targetFields = computed(() => targetSchema.value ? workflowInputFieldsFromSchema(targetSchema.value.schema_json) : [])

const sourceLabel = computed(() => sourceSchema.value
  ? `${sourceSchema.value.code} v${sourceSchema.value.version_number}`
  : t('common.loading'))

const targetLabel = computed(() => targetSchema.value
  ? `${targetSchema.value.code} v${targetSchema.value.version_number}`
  : t('workflowRouting.targetSchema'))

function listPath() {
  return workflowInputSchemaCompatibilitiesPath(orgId.value, workflowId.value, inputSchemaId.value)
}

async function loadData() {
  loading.value = true
  try {
    sourceSchema.value = await routing.fetchInputSchema(inputSchemaId.value)
    targetResults.value = await routing.searchInputSchemas(workflowId.value, '', inputSchemaId.value)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function searchTargets() {
  targetResults.value = await routing.searchInputSchemas(workflowId.value, targetQuery.value, inputSchemaId.value)
}

function selectTarget(schema: WorkflowInputSchemaRecord) {
  targetSchema.value = schema
  selectedTargetField.value = null
  fieldActions.value = {}
  defaultDrafts.value = Object.fromEntries(
    workflowInputFieldsFromSchema(schema.schema_json).map(field => [field.name, '']),
  )
  validationErrors.value = {}
}

function clearTargetSelection() {
  targetSchema.value = null
  selectedTargetField.value = null
  fieldActions.value = {}
  defaultDrafts.value = {}
  validationErrors.value = {}
  dragOverTargetName.value = null
}

function onDragField(field: WorkflowInputField) {
  draggedSourceField.value = field
}

function mapDraggedSource(target: WorkflowInputField) {
  if (!draggedSourceField.value) return
  dragOverTargetName.value = null
  if (!isDraggedFieldCompatible(target)) {
    return
  }
  const source = draggedSourceField.value
  const nextErrors = { ...validationErrors.value }
  delete nextErrors[target.name]
  validationErrors.value = nextErrors
  fieldActions.value = {
    ...fieldActions.value,
    [target.name]: { mode: 'source', sourcePath: source.name },
  }
  selectedTargetField.value = target
}

function showDropTarget(field: WorkflowInputField) {
  if (!draggedSourceField.value) return
  dragOverTargetName.value = field.name
  selectedTargetField.value = field
}

function hideDropTarget(field: WorkflowInputField) {
  if (dragOverTargetName.value === field.name) {
    dragOverTargetName.value = null
  }
}

function isDropTargetVisible(field: WorkflowInputField) {
  return dragOverTargetName.value === field.name && !!draggedSourceField.value
}

function isDraggedFieldCompatible(field: WorkflowInputField) {
  if (!draggedSourceField.value) return true
  const source = draggedSourceField.value
  return !source.type || !field.type || source.type === field.type
}

function actionLabel(field: WorkflowInputField) {
  const action = fieldActions.value[field.name]
  if (!action) return ''
  if (action.mode === 'source') return action.sourcePath
  if (action.mode === 'default') return `${t('workflowRouting.defaultValue')}: ${JSON.stringify(action.value)}`
  return ''
}

function setDefaultAction(field?: WorkflowInputField) {
  const target = field || selectedTargetField.value
  if (!target) return
  clearFieldAction(target)
  selectedTargetField.value = target
}

function clearFieldAction(field: WorkflowInputField) {
  const nextActions = { ...fieldActions.value }
  delete nextActions[field.name]
  fieldActions.value = nextActions
  const nextErrors = { ...validationErrors.value }
  delete nextErrors[field.name]
  validationErrors.value = nextErrors
}

function isSourceAction(field: WorkflowInputField) {
  return fieldActions.value[field.name]?.mode === 'source'
}

function parseDefaultDraft(fieldName: string): unknown {
  const raw = defaultDrafts.value[fieldName] ?? ''
  try {
    return JSON.parse(raw)
  }
  catch {
    return raw
  }
}

function buildPayloadFields(): CompatibilityMappingFieldPayload[] {
  return targetFields.value.map((field) => {
    const action = fieldActions.value[field.name]
    if (action?.mode === 'source') {
      return { target_path: field.name, source_path: action.sourcePath }
    }
    return { target_path: field.name, default: parseDefaultDraft(field.name) }
  })
}

async function submit() {
  if (!targetSchema.value) return
  saving.value = true
  validationErrors.value = {}
  try {
    await routing.validateAndCreateCompatibility(inputSchemaId.value, {
      target_input_schema_id: targetSchema.value.id,
      fields: buildPayloadFields(),
    })
    router.push(listPath())
  }
  catch (err) {
    const details = (err as Error & { details?: Array<{ field?: string, message?: string }> }).details
    if (Array.isArray(details)) {
      validationErrors.value = details.reduce<Record<string, string>>((acc, detail) => {
        if (detail.field) acc[detail.field] = detail.message || t('error.validation')
        return acc
      }, {})
    }
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateCompatibility')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <Button variant="ghost" class="-ml-3 mb-3 gap-2" @click="router.push(listPath())">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <div class="flex items-center gap-3">
          <div class="rounded-md bg-primary/10 p-3 text-primary">
            <Wand2 class="h-6 w-6" />
          </div>
          <div>
            <h1 class="text-2xl font-semibold">{{ t('workflowRouting.createCompatibility') }}</h1>
            <p class="text-sm text-muted-foreground">{{ sourceLabel }} -> {{ targetLabel }}</p>
          </div>
        </div>
      </div>
      <Button type="button" :disabled="saving || !targetSchema" data-testid="routing-compatibility-submit" @click="submit">
        {{ t('workflowRouting.validate') }}
      </Button>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>

    <div v-else class="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{{ t('workflowRouting.targetSchema') }}</CardTitle>
          <CardDescription>{{ t('workflowRouting.compatibilitiesDescription') }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div v-if="!targetSchema" class="space-y-4">
            <label class="block space-y-1.5">
              <span class="text-sm font-medium">{{ t('common.search') }}</span>
              <Input v-model="targetQuery" data-testid="routing-target-search" @input="searchTargets" />
            </label>
            <div class="flex flex-wrap gap-2">
              <Button
                v-for="schema in targetResults"
                :key="schema.id"
                type="button"
                size="sm"
                variant="outline"
                :data-testid="`routing-target-schema-${schema.id}`"
                @click="selectTarget(schema)"
              >
                {{ schema.code }} v{{ schema.version_number }}
              </Button>
            </div>
          </div>

          <div v-else class="flex items-center gap-3 rounded-md border bg-muted/20 p-3" data-testid="routing-selected-target-schema">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              class="h-8 w-8 shrink-0"
              aria-label="Clear target schema"
              data-testid="routing-clear-target-schema"
              @click="clearTargetSelection"
            >
              <X class="h-4 w-4" />
            </Button>
            <div class="min-w-0">
              <p class="truncate text-sm font-medium">
                {{ targetSchema.code }} v{{ targetSchema.version_number }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ t('workflowRouting.targetSchema') }}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <div v-if="targetSchema" class="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <RoutingSchemaFieldsColumn
          :title="t('workflowRouting.sourceSchema')"
          :fields="sourceFields"
          draggable
          @drag-field="onDragField"
        />

        <Card>
          <CardHeader>
            <CardTitle>{{ t('workflowRouting.compatibility') }}</CardTitle>
            <CardDescription>{{ targetLabel }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div
              v-for="field in targetFields"
              :key="field.name"
              class="grid gap-3 rounded-xl border bg-background p-3 md:grid-cols-[minmax(0,0.9fr)_minmax(0,1.3fr)]"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="truncate font-medium">{{ field.name }}</span>
                  <Badge variant="outline">{{ field.type }}</Badge>
                  <Badge v-if="field.required" variant="secondary">required</Badge>
                </div>
                <p v-if="field.description" class="mt-1 text-xs text-muted-foreground">
                  {{ field.description }}
                </p>
              </div>

              <div
                class="space-y-2 rounded-lg transition"
                @dragenter.prevent="showDropTarget(field)"
                @dragover.prevent="showDropTarget(field)"
                @dragleave="hideDropTarget(field)"
                @drop.prevent="mapDraggedSource(field)"
              >
                <div
                  v-if="isDropTargetVisible(field)"
                  class="flex min-h-11 w-full items-center justify-center rounded-md border border-dashed border-primary bg-primary/10 px-3 py-2 text-sm text-primary"
                  :class="isDraggedFieldCompatible(field) ? '' : 'cursor-not-allowed border-destructive bg-destructive/10 text-destructive'"
                  :data-testid="`routing-mapping-slot-${field.name}`"
                >
                  {{ draggedSourceField?.name || t('workflowRouting.sourceSchema') }}
                </div>

                <div
                  v-else-if="isSourceAction(field)"
                  class="flex min-h-11 w-full items-center justify-between rounded-md border bg-muted/20 px-3 py-2 text-left text-sm"
                  :class="validationErrors[field.name] ? 'border-destructive text-destructive' : ''"
                  :data-testid="`routing-mapping-slot-${field.name}`"
                >
                  <span class="truncate">{{ actionLabel(field) }}</span>
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    class="h-7 w-7 shrink-0"
                    aria-label="Clear mapping"
                    @click="clearFieldAction(field)"
                  >
                    <X class="h-4 w-4 text-muted-foreground" />
                  </Button>
                </div>

                <div v-else>
                  <Input
                    v-model="defaultDrafts[field.name]"
                    :placeholder="t('workflowRouting.defaultValue')"
                    :data-testid="`routing-default-value-${field.name}`"
                    @focus="selectedTargetField = field"
                  />
                </div>

                <p v-if="validationErrors[field.name]" class="text-xs text-destructive">
                  {{ validationErrors[field.name] }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import type { Node, Edge } from '@vue-flow/core'
import type { WorkflowInputSchemaField } from '~/composables/useWorkflows'
import type { StepData } from '~/composables/useDagEditor'
import type { WorkflowInputField } from './workflow-input-utils'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'
import { toast } from '~/components/ui/toast/use-toast'
import SchemaTree from './SchemaTree.vue'
import WorkflowInputsPanel from './WorkflowInputsPanel.vue'
import { workflowInputFieldsFromSchema, workflowInputPath } from './workflow-input-utils'

const props = defineProps<{
  stepId: string
  workflowId: number
  versionId: number | null
  allNodes: Node[]
  allEdges: Edge[]
}>()

const emit = defineEmits<{
  insertExpression: [expression: string]
  workflowInputsChanged: []
}>()

const { t } = useI18n()
const { fetchWorkflowInputSchema, upsertWorkflowInputSchemaField, deleteWorkflowInputSchemaField } = useWorkflows()
const searchQuery = ref('')
const workflowInputs = ref<WorkflowInputField[]>([])
const workflowInputSaving = ref(false)
const editingWorkflowInput = ref<WorkflowInputField | null>(null)
const deletingWorkflowInput = ref<WorkflowInputField | null>(null)
const workflowInputForm = ref<WorkflowInputSchemaField>({
  name: '',
  type: 'string',
  required: false,
  description: '',
})
const inputTypes: WorkflowInputSchemaField['type'][] = ['string', 'number', 'integer', 'boolean', 'object', 'array']
const canSaveWorkflowInput = computed(() => workflowInputForm.value.name.trim().length > 0 && !workflowInputSaving.value)

interface SourceNode {
  id: string
  label: string
  icon?: string
  color?: string
  outputSchema: Record<string, unknown>
}

function sourceNode(nodeId: string): SourceNode | null {
  const node = props.allNodes.find(n => n.id === nodeId)
  if (!node) return null
  const data = node.data as StepData
  if (!data.outputSchema) return null

  return {
    id: nodeId,
    label: data.label,
    icon: data.workTypeMeta?.icon,
    color: data.workTypeMeta?.color ?? '#607d8b',
    outputSchema: data.outputSchema,
  }
}

function collectSourceNodesThroughControl(
  nodeId: string,
  visited: Set<string>,
): SourceNode[] {
  if (visited.has(nodeId)) return []
  visited.add(nodeId)

  const node = props.allNodes.find(n => n.id === nodeId)
  if (!node) return []

  const data = node.data as StepData
  if (data.outputSchema && data.stepType !== 'control') {
    return [sourceNode(nodeId)].filter((node): node is SourceNode => Boolean(node))
  }

  const output: SourceNode[] = []
  for (const edge of props.allEdges.filter(e => e.target === nodeId)) {
    const upstreamSources = collectSourceNodesThroughControl(edge.source, visited)
    for (const upstreamSource of upstreamSources) {
      output.push(upstreamSource)
    }
  }
  return output
}

const directUpstreamNodes = computed<SourceNode[]>(() => {
  const upstreamEdges = props.allEdges.filter(e => e.target === props.stepId)
  const visited = new Set<string>()
  const collected = new Set<string>()
  const nodes: SourceNode[] = []

  for (const edge of upstreamEdges) {
    const upstreamNodes = collectSourceNodesThroughControl(edge.source, visited)
    for (const source of upstreamNodes) {
      if (collected.has(source.id)) continue
      collected.add(source.id)
      nodes.push(source)
    }
  }

  return nodes
})

const predecessorNodes = computed<SourceNode[]>(() => {
  const directIds = new Set(directUpstreamNodes.value.map(node => node.id))
  const visited = new Set<string>()
  const stack = [...directIds]

  while (stack.length > 0) {
    const current = stack.pop()
    if (!current) continue

    for (const edge of props.allEdges.filter(e => e.target === current)) {
      if (visited.has(edge.source) || directIds.has(edge.source) || edge.source === props.stepId) continue
      visited.add(edge.source)
      stack.push(edge.source)
    }
  }

  return [...visited]
    .map(sourceNode)
    .filter((node): node is SourceNode => Boolean(node))
})

function onFieldClick(path: string) {
  emit('insertExpression', path)
}

async function loadWorkflowInputs() {
  if (!props.workflowId) {
    workflowInputs.value = []
    return
  }
  const schema = await fetchWorkflowInputSchema(props.workflowId)
  workflowInputs.value = workflowInputFieldsFromSchema(schema)
}

async function createWorkflowInputFromField(
  field: string,
  property: Record<string, unknown>,
  onCreated: (expression: string) => void,
) {
  const type = typeof property.type === 'string' ? property.type : 'string'
  const allowedTypes: WorkflowInputSchemaField['type'][] = ['string', 'number', 'integer', 'boolean', 'object', 'array']
  await upsertWorkflowInputSchemaField(props.workflowId, {
    name: field,
    type: allowedTypes.includes(type as WorkflowInputSchemaField['type'])
      ? type as WorkflowInputSchemaField['type']
      : 'string',
    required: property.required === true,
    description: typeof property.description === 'string' ? property.description : undefined,
  })
  await loadWorkflowInputs()
  emit('workflowInputsChanged')
  onCreated(workflowInputPath(field))
}

function onEditWorkflowInput(input: WorkflowInputField) {
  deletingWorkflowInput.value = null
  editingWorkflowInput.value = input
  workflowInputForm.value = {
    name: input.name,
    type: input.type,
    required: input.required,
    description: input.description ?? '',
  }
}

function onDeleteWorkflowInput(input: WorkflowInputField) {
  editingWorkflowInput.value = null
  deletingWorkflowInput.value = input
}

async function saveWorkflowInput() {
  if (!canSaveWorkflowInput.value) return
  workflowInputSaving.value = true
  try {
    const description = workflowInputForm.value.description?.trim()
    await upsertWorkflowInputSchemaField(props.workflowId, {
      name: workflowInputForm.value.name.trim(),
      type: workflowInputForm.value.type,
      required: workflowInputForm.value.required,
      description: description || undefined,
    })
    await loadWorkflowInputs()
    editingWorkflowInput.value = null
    emit('workflowInputsChanged')
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    workflowInputSaving.value = false
  }
}

async function deleteWorkflowInput() {
  if (!deletingWorkflowInput.value || workflowInputSaving.value) return
  workflowInputSaving.value = true
  try {
    await deleteWorkflowInputSchemaField(props.workflowId, deletingWorkflowInput.value.name)
    await loadWorkflowInputs()
    deletingWorkflowInput.value = null
    emit('workflowInputsChanged')
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    workflowInputSaving.value = false
  }
}

watch(() => props.workflowId, loadWorkflowInputs, { immediate: true })

defineExpose({ createWorkflowInputFromField })
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.input') || 'Input' }}</h3>
    </div>

    <!-- Search -->
    <div class="px-4 py-2 shrink-0">
      <div class="relative">
        <Search class="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          placeholder="Search fields..."
          class="h-8 pl-8 text-xs"
        />
      </div>
    </div>

    <ScrollArea class="flex-1">
      <div class="px-4 pb-4 space-y-3">
        <div>
          <p class="mb-2 text-xs font-semibold text-muted-foreground">
            {{ t('nodeEditor.workflowInputs') }}
          </p>
          <WorkflowInputsPanel
            :inputs="workflowInputs"
            @insert-expression="onFieldClick"
            @edit-input="onEditWorkflowInput"
            @delete-input="onDeleteWorkflowInput"
          />
          <form
            v-if="editingWorkflowInput"
            class="mt-2 space-y-2 rounded-md border bg-muted/30 p-3"
            @submit.prevent="saveWorkflowInput"
          >
            <div class="grid grid-cols-[1fr_auto] gap-2">
              <label class="space-y-1">
                <span class="text-[11px] font-medium text-muted-foreground">{{ t('workflowInputs.name') }}</span>
                <Input v-model="workflowInputForm.name" readonly :disabled="workflowInputSaving" class="h-8 text-xs" />
              </label>
              <label class="space-y-1">
                <span class="text-[11px] font-medium text-muted-foreground">{{ t('workflowInputs.type') }}</span>
                <select
                  v-model="workflowInputForm.type"
                  class="flex h-8 rounded-md border border-input bg-background px-2 text-xs"
                  :disabled="workflowInputSaving"
                >
                  <option v-for="type in inputTypes" :key="type" :value="type">
                    {{ type }}
                  </option>
                </select>
              </label>
            </div>
            <label class="flex items-center gap-2 text-xs">
              <input v-model="workflowInputForm.required" type="checkbox" class="h-4 w-4 rounded border-input" :disabled="workflowInputSaving">
              <span>{{ t('workflowInputs.required') }}</span>
            </label>
            <label class="block space-y-1">
              <span class="text-[11px] font-medium text-muted-foreground">{{ t('workflowInputs.description') }}</span>
              <Input v-model="workflowInputForm.description" :disabled="workflowInputSaving" class="h-8 text-xs" />
            </label>
            <div class="flex justify-end gap-2">
              <Button type="button" size="sm" variant="outline" :disabled="workflowInputSaving" @click="editingWorkflowInput = null">
                {{ t('destructive.cancel') }}
              </Button>
              <Button type="submit" size="sm" :disabled="!canSaveWorkflowInput">
                {{ workflowInputSaving ? t('common.loading') : t('common.save') }}
              </Button>
            </div>
          </form>
          <div
            v-if="deletingWorkflowInput"
            class="mt-2 rounded-md border border-destructive/30 bg-destructive/5 p-3"
          >
            <p class="text-xs font-medium">
              {{ t('workflowInputs.deleteConfirmTitle', { name: deletingWorkflowInput.name }) }}
            </p>
            <p class="mt-1 text-xs text-muted-foreground">
              {{ t('workflowInputs.deleteConfirmBody') }}
            </p>
            <div class="mt-3 flex justify-end gap-2">
              <Button type="button" size="sm" variant="outline" :disabled="workflowInputSaving" @click="deletingWorkflowInput = null">
                {{ t('destructive.cancel') }}
              </Button>
              <Button
                type="button"
                size="sm"
                variant="destructive"
                :disabled="workflowInputSaving"
                data-testid="confirm-workflow-input-delete"
                @click="deleteWorkflowInput"
              >
                {{ workflowInputSaving ? t('common.loading') : t('destructive.delete') }}
              </Button>
            </div>
          </div>
          <Separator class="mt-3" />
        </div>

        <div v-if="directUpstreamNodes.length === 0 && predecessorNodes.length === 0" class="py-8 text-center">
          <p class="text-sm text-muted-foreground">
            {{ t('nodeEditor.noUpstream') || 'No upstream dependencies.' }}
          </p>
        </div>

        <div v-if="directUpstreamNodes.length > 0">
          <p class="mb-2 text-xs font-semibold text-muted-foreground">
            {{ t('nodeEditor.directUpstream') }}
          </p>
          <div v-for="upstream in directUpstreamNodes" :key="upstream.id">
            <div class="mb-2 flex items-center gap-2">
              <div
                class="h-2 w-2 rounded-full shrink-0"
                :style="{ backgroundColor: upstream.color }"
              />
              <span class="text-xs font-medium">
                {{ upstream.label }}
              </span>
              <span class="text-[10px] text-muted-foreground">
                #{{ upstream.id }}
              </span>
            </div>

            <SchemaTree
              :schema="upstream.outputSchema"
              :base-path="`$.steps.${upstream.id}.output`"
              :clickable="true"
              @field-click="onFieldClick"
            />

            <Separator class="mt-3" />
          </div>
        </div>

        <div v-if="predecessorNodes.length > 0">
          <p class="mb-2 text-xs font-semibold text-muted-foreground">
            {{ t('nodeEditor.earlierPredecessors') }}
          </p>
          <div v-for="upstream in predecessorNodes" :key="upstream.id">
            <div class="mb-2 flex items-center gap-2">
              <div
                class="h-2 w-2 rounded-full shrink-0"
                :style="{ backgroundColor: upstream.color }"
              />
              <span class="text-xs font-medium">
                {{ upstream.label }}
              </span>
              <span class="text-[10px] text-muted-foreground">
                #{{ upstream.id }}
              </span>
            </div>

            <SchemaTree
              :schema="upstream.outputSchema"
              :base-path="`$.steps.${upstream.id}.output`"
              :clickable="true"
              @field-click="onFieldClick"
            />

            <Separator class="mt-3" />
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

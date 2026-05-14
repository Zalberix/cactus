<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import type { Node, Edge } from '@vue-flow/core'
import type { WorkflowInputSchemaField } from '~/composables/useWorkflows'
import type { StepData } from '~/composables/useDagEditor'
import type { WorkflowInputField } from './workflow-input-utils'
import { Input } from '~/components/ui/input'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'
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
}>()

const { t } = useI18n()
const { fetchWorkflowInputSchema, upsertWorkflowInputSchemaField } = useWorkflows()
const searchQuery = ref('')
const workflowInputs = ref<WorkflowInputField[]>([])

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

const directUpstreamNodes = computed<SourceNode[]>(() => {
  const upstreamEdges = props.allEdges.filter(e => e.target === props.stepId)
  const nodes: SourceNode[] = []

  for (const edge of upstreamEdges) {
    const source = sourceNode(edge.source)
    if (source) nodes.push(source)
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
  onCreated(workflowInputPath(field))
}

watch(() => props.workflowId, loadWorkflowInputs, { immediate: true })

defineExpose({ createWorkflowInputFromField })
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.input') || 'Input' }}</h3>
      <p class="text-xs text-muted-foreground">
        {{ t('nodeEditor.upstream') || 'Upstream Outputs' }}
      </p>
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
          />
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

        <!-- Message context -->
        <div>
          <div class="mb-2 flex items-center gap-2">
            <div class="h-2 w-2 rounded-full shrink-0 bg-amber-500" />
            <span class="text-xs font-medium">Message</span>
          </div>
          <div
            class="flex items-center gap-1.5 rounded px-1 py-0.5 text-sm cursor-pointer hover:bg-accent"
            @click="onFieldClick('$.message.value')"
          >
            <span class="h-3 w-3 shrink-0" />
            <span class="inline-block h-2 w-2 rounded-full shrink-0 bg-gray-400" />
            <span class="font-medium">value</span>
            <span class="text-xs text-muted-foreground ml-auto">object</span>
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

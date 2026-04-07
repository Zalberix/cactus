import type { Node, Edge, Connection } from '@vue-flow/core'
import type { Step, Dependency } from '~/composables/useVersions'
import type { WorkTypeMeta } from '~/composables/useWorkers'

export interface StepData {
  label: string
  stepType: string
  workTypeId?: number
  workTypeName?: string
  workTypeCode?: string
  workTypeMeta?: WorkTypeMeta
  controlKind?: string
  config?: Record<string, unknown>
  controlSettings?: Record<string, unknown>
  inputMapping?: Record<string, string>
  settingsSchema?: Record<string, unknown>
  inputSchema?: Record<string, unknown>
  outputSchema?: Record<string, unknown>
  status: string
}

let positionTimer: ReturnType<typeof setTimeout> | null = null
const GRID_SIZE = 20

function snapToGrid(val: number): number {
  return Math.round(val / GRID_SIZE) * GRID_SIZE
}

export function useDagEditor(
  workflowId: Ref<number>,
  versionId: Ref<number | null>,
) {
  const {
    fetchSteps,
    createStep,
    updateStep,
    updateStepPosition,
    deleteStep: apiDeleteStep,
    createDependency,
    deleteDependency,
  } = useVersions()

  const nodes = ref<Node[]>([])
  const edges = ref<Edge[]>([])
  const selectedNodeId = ref<string | null>(null)
  const isDirty = ref(false)
  const validationErrors = ref<string[]>([])
  const isLoading = ref(false)

  const selectedNode = computed(() => {
    if (!selectedNodeId.value) return null
    return nodes.value.find(n => n.id === selectedNodeId.value) ?? null
  })

  function stepToNode(step: Step, fallbackIndex: number): Node {
    const pos = step.canvas_position ?? {
      x: fallbackIndex * 300,
      y: 200,
    }
    return {
      id: String(step.id),
      type: 'step',
      position: pos,
      data: {
        label: step.work_type_name ?? step.control_kind ?? `Step ${step.id}`,
        stepType: step.step_type,
        workTypeId: step.work_type_id,
        workTypeName: step.work_type_name,
        workTypeCode: step.work_type_code,
        workTypeMeta: step.work_type_meta,
        controlKind: step.control_kind,
        config: step.config,
        controlSettings: step.control_settings,
        inputMapping: step.input_mapping,
        settingsSchema: step.settings_schema,
        inputSchema: step.input_schema,
        outputSchema: step.output_schema,
        status: 'pending',
      } satisfies StepData,
    }
  }

  function dependencyToEdge(dep: Dependency): Edge {
    return {
      id: `e-${dep.depends_on_step_id}-${dep.step_id}-${dep.outcome}`,
      source: String(dep.depends_on_step_id),
      target: String(dep.step_id),
      sourceHandle: dep.outcome,
      label: dep.outcome,
      type: 'step',
    }
  }

  async function loadSteps(vid: number): Promise<void> {
    isLoading.value = true
    try {
      const result = await fetchSteps(vid)
      nodes.value = result.steps.map((step, index) => stepToNode(step, index))
      edges.value = result.dependencies.map(dependencyToEdge)
      isDirty.value = false
      validationErrors.value = []
    }
    finally {
      isLoading.value = false
    }
  }

  async function addStep(
    stepType: string,
    workTypeId: number | undefined,
    workTypeCode: string | undefined,
    position: { x: number; y: number },
    name?: string,
  ): Promise<void> {
    const vid = versionId.value
    if (!vid) return

    const stepName = name ?? `New ${stepType} step`
    const step = await createStep(vid, {
      name: stepName,
      step_type: stepType,
      work_type_id: stepType === 'task' ? workTypeId : undefined,
      control_kind: stepType === 'control' ? workTypeCode : undefined,
      canvas_position: position,
    })

    const node = stepToNode(step, nodes.value.length)
    node.position = position
    nodes.value = [...nodes.value, node]
    isDirty.value = true
  }

  async function removeStep(stepId: string): Promise<void> {
    await apiDeleteStep(Number(stepId))
    nodes.value = nodes.value.filter(n => n.id !== stepId)
    edges.value = edges.value.filter(
      e => e.source !== stepId && e.target !== stepId,
    )
    if (selectedNodeId.value === stepId) {
      selectedNodeId.value = null
    }
    isDirty.value = true
  }

  async function connectSteps(params: Connection): Promise<void> {
    if (!params.source || !params.target) return

    const outcome = params.sourceHandle ?? 'success'

    await createDependency(
      Number(params.target),
      Number(params.source),
      outcome,
    )

    const edge: Edge = {
      id: `e-${params.source}-${params.target}-${outcome}`,
      source: params.source,
      target: params.target,
      sourceHandle: outcome,
      label: outcome,
      type: 'step',
    }
    edges.value = [...edges.value, edge]
    isDirty.value = true
  }

  async function removeEdge(edgeId: string): Promise<void> {
    const edge = edges.value.find(e => e.id === edgeId)
    if (!edge) return

    await deleteDependency(Number(edge.target), Number(edge.source))
    edges.value = edges.value.filter(e => e.id !== edgeId)
    isDirty.value = true
  }

  function updateNodeData(nodeId: string, data: Partial<StepData>): void {
    nodes.value = nodes.value.map((n) => {
      if (n.id !== nodeId) return n
      return {
        ...n,
        data: { ...n.data, ...data },
      }
    })
    isDirty.value = true
  }

  async function saveVersion(): Promise<boolean> {
    const vid = versionId.value
    if (!vid) return false

    const { validateVersion } = useVersions()
    const result = await validateVersion(vid)

    if (!result.valid && result.errors?.length) {
      validationErrors.value = result.errors
      return false
    }

    validationErrors.value = []
    isDirty.value = false
    return true
  }

  function onNodeDragStop(nodeId: string, position: { x: number; y: number }): void {
    const snapped = {
      x: snapToGrid(position.x),
      y: snapToGrid(position.y),
    }
    if (positionTimer) clearTimeout(positionTimer)
    positionTimer = setTimeout(() => {
      updateStepPosition(Number(nodeId), snapped).catch(() => {})
    }, 300)
  }

  function selectNode(nodeId: string | null): void {
    selectedNodeId.value = nodeId
  }

  async function updateStepOnServer(stepId: string, data: Partial<Step>): Promise<void> {
    await updateStep(Number(stepId), data)
    isDirty.value = true
  }

  return {
    nodes,
    edges,
    selectedNodeId,
    selectedNode,
    isDirty,
    validationErrors,
    isLoading,
    loadSteps,
    addStep,
    removeStep,
    connectSteps,
    removeEdge,
    updateNodeData,
    saveVersion,
    onNodeDragStop,
    selectNode,
    updateStepOnServer,
  }
}

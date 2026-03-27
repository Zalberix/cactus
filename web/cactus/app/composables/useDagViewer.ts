import type { Node, Edge } from '@vue-flow/core'
import type { WsStepStatus } from './useWebSocketStatus'
import type { StepStatus } from './useMessages'

const NODE_WIDTH = 200
const NODE_HEIGHT = 60
const NODE_GAP_X = 40
const NODE_GAP_Y = 80

export function useDagViewer(messageId: Ref<number>) {
  const { fetchMessageStatus } = useMessages()

  const {
    steps: wsSteps,
    workflowStatus,
    isConnected,
    error,
  } = useWebSocketStatus(messageId)

  const staticSteps = ref<StepStatus[]>([])
  const selectedStep = ref<number | null>(null)
  const initialLoading = ref(true)

  // Fetch initial status from REST API for the step list structure
  async function loadInitialStatus() {
    try {
      const status = await fetchMessageStatus(messageId.value)
      staticSteps.value = status.steps
    }
    catch {
      // Will rely on WebSocket snapshot
    }
    finally {
      initialLoading.value = false
    }
  }

  // Load saved positions from localStorage
  function loadPositions(): Record<string, { x: number; y: number }> {
    try {
      const saved = localStorage.getItem(`dag-viewer-${messageId.value}`)
      if (saved) {
        return JSON.parse(saved)
      }
    }
    catch {
      // Ignore parse errors
    }
    return {}
  }

  // Generate default grid layout positions
  function defaultPosition(index: number, total: number): { x: number; y: number } {
    const cols = Math.max(1, Math.ceil(Math.sqrt(total)))
    const row = Math.floor(index / cols)
    const col = index % cols

    return {
      x: col * (NODE_WIDTH + NODE_GAP_X) + 50,
      y: row * (NODE_HEIGHT + NODE_GAP_Y) + 50,
    }
  }

  // Merge static step data with live WS status
  function getMergedStep(stepId: number): WsStepStatus | StepStatus | undefined {
    const wsStep = wsSteps.value.get(stepId)
    if (wsStep) return wsStep

    return staticSteps.value.find(s => s.step_id === stepId)
  }

  const nodes = computed<Node[]>(() => {
    const savedPositions = loadPositions()

    // Build from WS steps if available, otherwise from static
    const stepList: Array<{ step_id: number; step_type: string }> = []

    if (wsSteps.value.size > 0) {
      for (const [, step] of wsSteps.value) {
        stepList.push({ step_id: step.step_id, step_type: step.step_type })
      }
    }
    else {
      for (const step of staticSteps.value) {
        stepList.push({ step_id: step.step_id, step_type: step.step_type })
      }
    }

    return stepList.map((step, index) => {
      const merged = getMergedStep(step.step_id)
      const status = merged?.status ?? 'pending'
      const nodeId = String(step.step_id)

      const staticStep = staticSteps.value.find(s => s.step_id === step.step_id)
      const label = `Step ${step.step_id}`

      return {
        id: nodeId,
        type: 'step',
        position: savedPositions[nodeId] ?? defaultPosition(index, stepList.length),
        data: {
          label,
          stepType: step.step_type,
          status,
          error_message: (merged as StepStatus)?.error_message
            ?? (merged as WsStepStatus)?.error,
          started_at: (merged as StepStatus)?.started_at
            ?? (merged as WsStepStatus)?.started_at,
          completed_at: (merged as StepStatus)?.completed_at
            ?? (merged as WsStepStatus)?.completed_at,
        },
      } satisfies Node
    })
  })

  // Build edges from step dependencies (based on step order for now)
  // The real edges would come from the workflow DAG definition,
  // but the status API only returns step status, not structure.
  // For the viewer, steps are arranged as a sequence.
  const edges = computed<Edge[]>(() => {
    const edgeList: Edge[] = []

    const stepIds: number[] = []
    if (wsSteps.value.size > 0) {
      for (const [id] of wsSteps.value) {
        stepIds.push(id)
      }
    }
    else {
      for (const step of staticSteps.value) {
        stepIds.push(step.step_id)
      }
    }

    stepIds.sort((a, b) => a - b)

    for (let i = 0; i < stepIds.length - 1; i++) {
      edgeList.push({
        id: `e-${stepIds[i]}-${stepIds[i + 1]}`,
        source: String(stepIds[i]),
        target: String(stepIds[i + 1]),
        sourceHandle: 'success',
        type: 'step',
      })
    }

    return edgeList
  })

  onMounted(() => {
    loadInitialStatus()
  })

  return {
    nodes,
    edges,
    workflowStatus,
    isConnected,
    error,
    selectedStep,
    initialLoading,
  }
}

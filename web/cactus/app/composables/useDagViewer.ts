import type { Edge, Node } from '@vue-flow/core'
import type { GraphStep, MessageDetail, StepRunDetail } from './useMessages'

const NODE_WIDTH = 200
const NODE_HEIGHT = 60
const NODE_GAP_X = 40
const NODE_GAP_Y = 80
const DEFAULT_COLUMNS = 4

export function normalizeRuntimeStatus(status?: string): string {
  const normalized = status ?? 'pending'
  if (normalized === 'done') return 'completed'
  if (normalized === 'error') return 'failed'
  if (['pending', 'running', 'completed', 'failed', 'skipped'].includes(normalized)) {
    return normalized
  }
  return 'pending'
}

export function toCanvasPosition(
  raw: Record<string, unknown> | undefined,
  index: number,
): { x: number; y: number } {
  if (typeof raw?.x === 'number' && Number.isFinite(raw.x)
    && typeof raw?.y === 'number' && Number.isFinite(raw.y)) {
    return { x: raw.x, y: raw.y }
  }

  const col = index % DEFAULT_COLUMNS
  const row = Math.floor(index / DEFAULT_COLUMNS)
  return {
    x: col * (NODE_WIDTH + NODE_GAP_X) + 50,
    y: row * (NODE_HEIGHT + NODE_GAP_Y) + 50,
  }
}

export function runtimeStepMap(steps: readonly StepRunDetail[] = []): Map<number, StepRunDetail> {
  const map = new Map<number, StepRunDetail>()
  for (const step of steps) {
    map.set(step.step_id, step)
  }
  return map
}

export type RuntimeEdgeState = 'pending' | 'running' | 'completed' | 'failed' | 'skipped'

export function runtimeEdgeState(input: {
  sourceStatus?: string
  targetStatus?: string
  sourceOutcome?: string
  edgeOutcome?: string
}): RuntimeEdgeState {
  const sourceStatus = normalizeRuntimeStatus(input.sourceStatus)
  const targetStatus = normalizeRuntimeStatus(input.targetStatus)
  const edgeOutcome = input.edgeOutcome || 'success'
  const sourceOutcome = input.sourceOutcome || edgeOutcome
  const outcomeMatches = sourceOutcome === edgeOutcome

  if (sourceStatus === 'failed' || targetStatus === 'failed') return 'failed'
  if (!outcomeMatches) return 'pending'
  if (sourceStatus === 'skipped' || targetStatus === 'skipped') return 'skipped'
  if (sourceStatus !== 'completed') return 'pending'
  if (targetStatus === 'running') return 'running'
  return 'completed'
}

export function useDagViewer(messageId: Ref<number>) {
  const { fetchMessageDetail } = useMessages()

  const {
    detail: wsDetail,
    runSteps: wsRunSteps,
    workflowStatus,
    isConnected,
    error,
  } = useWebSocketStatus(messageId)

  const restDetail = ref<MessageDetail | null>(null)
  const selectedStep = ref<number | null>(null)
  const initialLoading = ref(true)

  const detail = computed(() => wsDetail.value ?? restDetail.value)
  const runSteps = computed(() => {
    if (wsRunSteps.value.size > 0) return wsRunSteps.value
    return runtimeStepMap(detail.value?.run_steps ?? [])
  })

function stepDisplayName(step: Pick<GraphStep, 'id' | 'name' | 'work_type_name' | 'control_kind'>): string {
  if (step.control_kind === 'start' && !step.name) return 'System Trigger'
  return step.name ?? step.work_type_name ?? step.control_kind ?? `Step ${step.id}`
}

  async function loadInitialDetail() {
    try {
      restDetail.value = await fetchMessageDetail(messageId.value)
    }
    catch {
      // WebSocket snapshot can still populate the same detail contract.
    }
    finally {
      initialLoading.value = false
    }
  }

  const nodes = computed<Node[]>(() =>
    (detail.value?.graph.steps ?? []).map((step, index) => {
      const runtime = runSteps.value.get(step.id)
      return {
        id: String(step.id),
        type: 'step',
        position: toCanvasPosition(step.canvas_position, index),
        draggable: false,
        data: {
          label: stepDisplayName(step),
          stepType: step.step_type,
          controlKind: step.control_kind,
          workTypeName: step.work_type_name,
          workTypeCode: step.work_type_code,
          workTypeMeta: step.work_type_meta,
          inputMapping: [...(step.input_mapping ?? [])],
          inputSchema: step.input_schema,
          outputSchema: step.output_schema,
          status: normalizeRuntimeStatus(runtime?.status),
          error_message: runtime?.error_message,
          started_at: runtime?.started_at,
          completed_at: runtime?.completed_at,
        },
      } satisfies Node
    }),
  )

  const edges = computed<Edge[]>(() =>
    (detail.value?.graph.dependencies ?? []).map((dep) => {
      const sourceRuntime = runSteps.value.get(dep.depends_on_step_id)
      const targetRuntime = runSteps.value.get(dep.step_id)
      const outcome = dep.outcome || 'success'

      return {
        id: `e-${dep.depends_on_step_id}-${dep.step_id}-${outcome}-${dep.output_index}`,
        source: String(dep.depends_on_step_id),
        target: String(dep.step_id),
        sourceHandle: outcome,
        type: 'step',
        data: {
          runtimeState: runtimeEdgeState({
            sourceStatus: sourceRuntime?.status,
            targetStatus: targetRuntime?.status,
            sourceOutcome: sourceRuntime?.outcome,
            edgeOutcome: outcome,
          }),
        },
      } satisfies Edge
    }),
  )

  onMounted(() => {
    loadInitialDetail()
  })

  return {
    detail,
    runSteps,
    nodes,
    edges,
    workflowStatus,
    isConnected,
    error,
    selectedStep,
    initialLoading,
  }
}

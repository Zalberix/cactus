import { toast } from '~/components/ui/toast/use-toast'
import type { MessageDetail, StepRunDetail } from './useMessages'

export interface WsStepStatus {
  step_id: number
  step_type: string
  status: string
  started_at?: string
  completed_at?: string
  duration_ms?: number
  error?: string
}

export function useWebSocketStatus(messageId: Ref<number>) {
  const authStore = useAuthStore()
  const { t } = useI18n()

  const detail = ref<MessageDetail | null>(null)
  const runSteps = ref<Map<number, StepRunDetail>>(new Map())
  const terminalStatus = ref<string | null>(null)
  const workflowStatus = computed(() => detail.value?.workflow_run?.status ?? terminalStatus.value ?? 'pending')
  const isConnected = ref(false)
  const error = ref<string | null>(null)

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  let isTerminal = false

  const MAX_RECONNECT_DELAY = 30000

  function getReconnectDelay(): number {
    const base = 3000
    const delay = base * Math.pow(2, reconnectAttempts)
    return Math.min(delay, MAX_RECONNECT_DELAY)
  }

  function rebuildRunSteps(steps: StepRunDetail[] = []) {
    const next = new Map<number, StepRunDetail>()
    for (const step of steps) {
      next.set(step.step_id, step)
    }
    runSteps.value = next
  }

  function connect() {
    if (isTerminal) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/workflow/${messageId.value}`

    try {
      ws = new WebSocket(wsUrl)
    }
    catch (err) {
      error.value = String(err)
      isConnected.value = false
      scheduleReconnect()
      return
    }

    ws.onopen = () => {
      ws!.send(JSON.stringify({
        type: 'auth',
        token: authStore.accessToken,
      }))
    }

    ws.onmessage = (event: MessageEvent) => {
      let msg: Record<string, unknown>
      try {
        msg = JSON.parse(event.data as string)
      }
      catch {
        return
      }

      switch (msg.type) {
        case 'auth_ok':
          isConnected.value = true
          reconnectAttempts = 0
          error.value = null
          break

        case 'snapshot':
          detail.value = msg.detail as MessageDetail
          terminalStatus.value = null
          rebuildRunSteps(detail.value?.run_steps ?? [])
          break

        case 'step_update': {
          const update = msg as {
            step_id: number
            run_step_id?: number
            status?: string
            input_data?: Record<string, unknown>
            output_data?: Record<string, unknown>
            started_at?: string
            completed_at?: string
            duration_ms?: number
            error?: string
          }
          const existing = runSteps.value.get(update.step_id) ?? {
            id: update.run_step_id ?? 0,
            step_id: update.step_id,
            status: update.status ?? 'pending',
          }
          runSteps.value.set(update.step_id, {
            ...existing,
            id: update.run_step_id ?? existing.id,
            status: update.status ?? existing.status,
            input_data: update.input_data ?? existing.input_data,
            output_data: update.output_data ?? existing.output_data,
            started_at: update.started_at ?? existing.started_at,
            completed_at: update.completed_at ?? existing.completed_at,
            duration_ms: update.duration_ms ?? existing.duration_ms,
            error_message: update.error ?? existing.error_message,
          })
          runSteps.value = new Map(runSteps.value)
          break
        }

        case 'workflow_done':
          terminalStatus.value = 'completed'
          if (detail.value?.workflow_run) detail.value.workflow_run.status = 'completed'
          isTerminal = true
          toast({ title: t('messages.detail.workflowCompleted') })
          break

        case 'workflow_failed':
          terminalStatus.value = 'failed'
          if (detail.value?.workflow_run) detail.value.workflow_run.status = 'failed'
          error.value = (msg.error as string) ?? null
          isTerminal = true
          toast({
            title: t('messages.detail.workflowFailed'),
            variant: 'destructive',
          })
          break
      }
    }

    ws.onclose = () => {
      isConnected.value = false
      if (!isTerminal) {
        toast({
          title: t('error.wsDisconnected'),
          variant: 'destructive',
        })
        scheduleReconnect()
      }
    }

    ws.onerror = () => {
      isConnected.value = false
    }
  }

  function scheduleReconnect() {
    if (isTerminal) return
    if (reconnectTimer) return

    const delay = getReconnectDelay()
    reconnectAttempts++

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  function disconnect() {
    isTerminal = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    detail: readonly(detail),
    runSteps: readonly(runSteps),
    workflowStatus: readonly(workflowStatus),
    isConnected: readonly(isConnected),
    error: readonly(error),
  }
}

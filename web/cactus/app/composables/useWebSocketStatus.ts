import { toast } from '~/components/ui/toast/use-toast'

export interface WsStepStatus {
  step_id: number
  step_type: string
  status: string
  started_at?: string
  completed_at?: string
  error?: string
}

export function useWebSocketStatus(messageId: Ref<number>) {
  const authStore = useAuthStore()
  const { t } = useI18n()

  const steps = ref<Map<number, WsStepStatus>>(new Map())
  const workflowStatus = ref<string>('pending')
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
          workflowStatus.value = msg.workflow_status as string
          steps.value = new Map()
          if (Array.isArray(msg.steps)) {
            for (const s of msg.steps) {
              steps.value.set(s.step_id as number, {
                step_id: s.step_id as number,
                step_type: s.step_type as string,
                status: s.status as string,
                started_at: s.started_at as string | undefined,
                completed_at: s.completed_at as string | undefined,
              })
            }
          }
          // Trigger reactivity by reassigning the Map
          steps.value = new Map(steps.value)
          break

        case 'step_update': {
          const existing = steps.value.get(msg.step_id as number)
          steps.value.set(msg.step_id as number, {
            step_id: msg.step_id as number,
            step_type: (existing?.step_type ?? msg.step_type ?? '') as string,
            status: msg.status as string,
            started_at: (msg.started_at ?? existing?.started_at) as string | undefined,
            completed_at: (msg.completed_at ?? existing?.completed_at) as string | undefined,
            error: (msg.error ?? existing?.error) as string | undefined,
          })
          // Trigger reactivity
          steps.value = new Map(steps.value)
          break
        }

        case 'workflow_done':
          workflowStatus.value = 'completed'
          isTerminal = true
          toast({ title: t('messages.detail.workflowCompleted') })
          break

        case 'workflow_failed':
          workflowStatus.value = 'failed'
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
    steps: readonly(steps),
    workflowStatus: readonly(workflowStatus),
    isConnected: readonly(isConnected),
    error: readonly(error),
  }
}

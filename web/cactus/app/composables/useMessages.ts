import type { ApiResponse, PaginationMeta } from '~/utils/api-types'

export interface MessageListItem {
  id: number
  workflow_id: number
  workflow_name?: string
  status: string
  created_at: string
  updated_at: string
}

export interface StepStatus {
  id: number
  step_id: number
  step_type: string
  status: string
  outcome?: string
  started_at?: string
  completed_at?: string
  error_message?: string
}

export interface WorkflowRunStatus {
  id: number
  status: string
  started_at?: string
  completed_at?: string
  error_message?: string
}

export interface MessageStatus {
  message_id: number
  message_status: string
  created_at: string
  workflow_run?: WorkflowRunStatus
  steps: StepStatus[]
}

export interface MessageDetail {
  message_id: number
  workflow_id: number
  workflow_name: string
  message_status: string
  message_value: Record<string, unknown>
  created_at: string
  updated_at: string
  workflow_run?: WorkflowRunStatus
  graph: MessageGraph
  run_steps: StepRunDetail[]
}

export interface MessageGraph {
  version_id: number
  steps: GraphStep[]
  dependencies: GraphDependency[]
}

export interface GraphStep {
  id: number
  step_type: string
  control_kind?: string
  work_type_id?: number
  work_type_name?: string
  work_type_code?: string
  work_type_meta?: Record<string, unknown>
  input_mapping: MappingEntry[]
  canvas_position?: Record<string, unknown>
  input_schema?: Record<string, unknown>
  output_schema?: Record<string, unknown>
}

export interface MappingEntry {
  target: string
  source: string
}

export interface GraphDependency {
  step_id: number
  depends_on_step_id: number
  outcome: string
  output_index: number
}

export interface StepRunDetail {
  id: number
  step_id: number
  status: string
  outcome?: string
  input_data?: Record<string, unknown>
  output_data?: Record<string, unknown>
  started_at?: string
  completed_at?: string
  error_message?: string
}

export function useMessages() {
  const { api } = useApi()

  async function fetchMessageStatus(messageId: number): Promise<MessageStatus> {
    const resp = await api<ApiResponse<MessageStatus>>(
      `/messages/${messageId}/status`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch message status')
    }
    return resp.data
  }

  async function fetchMessageDetail(messageId: number): Promise<MessageDetail> {
    const resp = await api<ApiResponse<MessageDetail>>(
      `/messages/${messageId}/detail`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch message detail')
    }
    return resp.data
  }

  async function fetchMessages(
    orgId: number,
    page: number = 1,
    perPage: number = 20,
  ): Promise<{ data: MessageListItem[]; meta: PaginationMeta }> {
    const resp = await api<ApiResponse<MessageListItem[]>>(
      `/organizations/${orgId}/messages`,
      {
        params: { page, per_page: perPage },
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch messages')
    }
    return {
      data: resp.data,
      meta: resp.meta ?? { total: 0, page, per_page: perPage, total_pages: 0 },
    }
  }

  return {
    fetchMessageDetail,
    fetchMessageStatus,
    fetchMessages,
  }
}

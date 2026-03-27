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

  async function fetchMessages(
    _orgId: number,
    _page: number = 1,
    perPage: number = 20,
  ): Promise<{ data: MessageListItem[]; meta: PaginationMeta }> {
    // TODO: Backend endpoint GET /organizations/:orgId/messages not yet implemented
    // See RESEARCH.md Open Question #1
    console.warn('[useMessages] ListMessages endpoint not implemented in backend')
    return {
      data: [] as MessageListItem[],
      meta: { total: 0, page: 1, per_page: perPage, total_pages: 0 },
    }
  }

  return {
    fetchMessageStatus,
    fetchMessages,
  }
}

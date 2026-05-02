import type { ApiResponse } from './api-types'

type ErrorPayload = {
  data?: unknown
  response?: {
    _data?: unknown
  }
  statusMessage?: string
  message?: string
}

function extractApiMessage(payload: unknown): string | null {
  if (!payload || typeof payload !== 'object') return null

  const response = payload as Partial<ApiResponse<unknown>>
  return response.error?.message || null
}

export function getErrorMessage(error: unknown, fallback: string): string {
  const fetchError = error as ErrorPayload

  return (
    extractApiMessage(error)
    || extractApiMessage(fetchError.data)
    || extractApiMessage(fetchError.response?._data)
    || fetchError.statusMessage
    || (error instanceof Error ? error.message : null)
    || (typeof error === 'string' ? error : null)
    || fallback
  )
}

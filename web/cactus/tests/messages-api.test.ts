import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useMessages } from '../app/composables/useMessages'

describe('messages API composable', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('parses message list routing fields', async () => {
    const calls: Array<[string, unknown]> = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string, options?: unknown) => {
        calls.push([url, options])
        return {
          success: true,
          data: [{
            id: 101,
            workflow_id: 7,
            workflow_name: 'Welcome',
            workflow_version_id: 31,
            workflow_version_number: 4,
            workflow_version_name: 'Canary',
            workflow_input_schema_id: 11,
            workflow_input_schema_code: 'public',
            workflow_input_schema_version: 3,
            input_schema_compatibility_id: 41,
            experiment_id: 51,
            experiment_scope_id: 61,
            experiment_variant_id: 71,
            selection_reason: 'experiment',
            status: 'running',
            created_at: '2026-06-04T10:00:00Z',
            updated_at: '2026-06-04T10:00:01Z',
          }],
          meta: { total: 1, page: 2, per_page: 10, total_pages: 1 },
        }
      },
    }))

    const { fetchMessages } = useMessages()
    const result = await fetchMessages(9, 2, 10)

    expect(calls).toEqual([[
      '/organizations/9/messages',
      { params: { page: 2, per_page: 10 } },
    ]])
    expect(result.data[0]).toMatchObject({
      workflow_input_schema_id: 11,
      workflow_input_schema_code: 'public',
      workflow_input_schema_version: 3,
      input_schema_compatibility_id: 41,
      experiment_id: 51,
      experiment_scope_id: 61,
      experiment_variant_id: 71,
            selection_reason: 'experiment',
    })
    expect(result.meta).toEqual({ total: 1, page: 2, per_page: 10, total_pages: 1 })
  })

  it('parses message detail public input and mapped version input', async () => {
    vi.stubGlobal('useApi', () => ({
      api: async () => ({
        success: true,
        data: {
          message_id: 101,
          workflow_id: 7,
          workflow_name: 'Welcome',
          workflow_version_id: 31,
          workflow_version_number: 4,
          workflow_version_name: 'Canary',
          workflow_input_schema_id: 11,
          workflow_input_schema_code: 'public',
          workflow_input_schema_version: 3,
          input_schema_compatibility_id: 41,
          experiment_id: 51,
          experiment_scope_id: 61,
          experiment_variant_id: 71,
          selection_reason: 'experiment',
          version_input_data: { recipient: { email: 'ada@example.com' } },
          routing_decision: { reason: 'experiment_variant', workflow_version_id: 31 },
          message_status: 'running',
          message_value: { email: 'ada@example.com' },
          created_at: '2026-06-04T10:00:00Z',
          updated_at: '2026-06-04T10:00:01Z',
          graph: { version_id: 31, steps: [], dependencies: [] },
          run_steps: [],
        },
      }),
    }))

    const { fetchMessageDetail } = useMessages()
    const detail = await fetchMessageDetail(101)

    expect(detail.message_value).toEqual({ email: 'ada@example.com' })
    expect(detail.version_input_data).toEqual({ recipient: { email: 'ada@example.com' } })
    expect(detail.routing_decision).toEqual({ reason: 'experiment_variant', workflow_version_id: 31 })
    expect(detail.input_schema_compatibility_id).toBe(41)
    expect(detail.experiment_variant_id).toBe(71)
    expect(detail.selection_reason).toBe('experiment')
  })
})

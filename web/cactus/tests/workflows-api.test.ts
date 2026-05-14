import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useWorkflows } from '../app/composables/useWorkflows'

describe('workflow API composable', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('deletes workflow input schema fields through the field endpoint', async () => {
    const calls: Array<[string, unknown]> = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string, options?: unknown) => {
        calls.push([url, options])
        return {
          success: true,
          data: { schema: { type: 'object', properties: {} } },
        }
      },
    }))

    const { deleteWorkflowInputSchemaField } = useWorkflows()
    const schema = await deleteWorkflowInputSchemaField(42, 'email')

    expect(calls).toEqual([
      ['/workflows/42/input-schema/fields/email', { method: 'DELETE' }],
    ])
    expect(schema).toEqual({ type: 'object', properties: {} })
  })
})

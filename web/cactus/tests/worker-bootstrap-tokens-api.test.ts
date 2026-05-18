import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useWorkers } from '../app/composables/useWorkers'

describe('worker bootstrap token API composable', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('creates org-scoped worker bootstrap tokens', async () => {
    const calls: Array<[string, unknown]> = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string, options?: unknown) => {
        calls.push([url, options])
        return { success: true, data: { token: { id: 1 }, plaintext: 'plain' } }
      },
    }))

    const { createWorkerBootstrapToken } = useWorkers()
    const result = await createWorkerBootstrapToken(12, {
      name: 'smtp-prod',
      work_type_id: 3,
      max_active_workers: 10,
      expires_at: '2026-06-16T00:00:00Z',
      description: 'SMTP pool',
    })

    expect(calls).toEqual([
      ['/organizations/12/worker-bootstrap-tokens', {
        method: 'POST',
        body: {
          name: 'smtp-prod',
          work_type_id: 3,
          max_active_workers: 10,
          expires_at: '2026-06-16T00:00:00Z',
          description: 'SMTP pool',
        },
      }],
    ])
    expect(result.plaintext).toBe('plain')
  })
})

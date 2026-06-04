import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useWorkflowRouting } from '../app/composables/useWorkflowRouting'

describe('workflow routing API composable', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('parses input schema list responses', async () => {
    const calls: Array<[string, unknown]> = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string, options?: unknown) => {
        calls.push([url, options])
        return {
          success: true,
          data: [{
            id: 11,
            workflow_id: 7,
            code: 'public',
            version_number: 3,
            schema_json: { type: 'object', properties: { email: { type: 'string' } } },
            status: 'active',
            is_default: true,
          }],
        }
      },
    }))

    const { fetchInputSchemas } = useWorkflowRouting()
    const schemas = await fetchInputSchemas(7)

    expect(calls).toEqual([['/workflows/7/input-schemas', undefined]])
    expect(schemas).toEqual([{
      id: 11,
      workflow_id: 7,
      code: 'public',
      version_number: 3,
      schema_json: { type: 'object', properties: { email: { type: 'string' } } },
      status: 'active',
      is_default: true,
    }])
  })

  it('normalizes compatibility pgtype fields', async () => {
    vi.stubGlobal('useApi', () => ({
      api: async () => ({
        success: true,
        data: [{
          id: 21,
          workflow_version_id: 31,
          workflow_input_schema_id: 41,
          compatibility_type: 'adapter',
          workflow_input_mapper_id: { Int32: 51, Valid: true },
          default_values: { locale: 'en' },
          is_active: true,
          is_default_route: false,
        }, {
          id: 22,
          workflow_version_id: 32,
          workflow_input_schema_id: 42,
          compatibility_type: 'native',
          workflow_input_mapper_id: { Int32: 0, Valid: false },
          default_values: {},
          is_active: true,
          is_default_route: true,
        }],
      }),
    }))

    const { fetchCompatibilities } = useWorkflowRouting()
    const compatibilities = await fetchCompatibilities(7)

    expect(compatibilities[0].workflow_input_mapper_id).toBe(51)
    expect(compatibilities[0].default_values).toEqual({ locale: 'en' })
    expect(compatibilities[1].workflow_input_mapper_id).toBeUndefined()
    expect(compatibilities[1].is_default_route).toBe(true)
  })

  it('normalizes experiment timestamp and text fields', async () => {
    vi.stubGlobal('useApi', () => ({
      api: async () => ({
        success: true,
        data: [{
          id: 61,
          workflow_id: 7,
          name: 'Canary',
          description: { String: 'Ten percent rollout', Valid: true },
          experiment_type: 'canary',
          status: 'active',
          started_at: { Time: '2026-06-04T10:00:00Z', Valid: true },
          ended_at: { Time: '', Valid: false },
        }],
      }),
    }))

    const { fetchExperiments } = useWorkflowRouting()
    const experiments = await fetchExperiments(7)

    expect(experiments).toEqual([{
      id: 61,
      workflow_id: 7,
      name: 'Canary',
      description: 'Ten percent rollout',
      experiment_type: 'canary',
      status: 'active',
      started_at: '2026-06-04T10:00:00Z',
      ended_at: undefined,
    }])
  })

  it('normalizes experiment scope fallback version ids', async () => {
    vi.stubGlobal('useApi', () => ({
      api: async () => ({
        success: true,
        data: [{
          id: 71,
          workflow_experiment_id: 61,
          workflow_input_schema_id: 11,
          traffic_conditions: { fields: { country: { eq: 'US' } } },
          conditions_hash: 'abc',
          traffic_percent: 25,
          fallback_policy: 'fallback_version',
          fallback_workflow_version_id: { Int32: 31, Valid: true },
        }],
      }),
    }))

    const { fetchExperimentScopes } = useWorkflowRouting()
    const scopes = await fetchExperimentScopes(61)

    expect(scopes[0].fallback_workflow_version_id).toBe(31)
    expect(scopes[0].traffic_conditions).toEqual({ fields: { country: { eq: 'US' } } })
  })
})

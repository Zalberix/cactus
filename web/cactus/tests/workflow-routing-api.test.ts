import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { PaginationMeta } from '../app/utils/api-types'
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
      usage: {
        used_by_message: false,
        used_by_mapper: false,
        used_by_experiment: false,
        is_readonly: false,
        reasons: [],
      },
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

  it('normalizes routing rows as input schema versions with supported workflow versions', async () => {
    const calls: string[] = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string) => {
        calls.push(url)
        return {
          success: true,
          data: [{
            input_schema_id: 11,
            workflow_id: 7,
            input_schema_code: 'v2',
            input_schema_version_number: 2,
            input_schema_status: 'active',
            is_default: true,
            supported_versions: [{
              compatibility_id: 31,
              workflow_version_id: 21,
              workflow_version_name: 'Version 1',
              workflow_version_number: 1,
              compatibility_type: 'native',
              workflow_input_mapper_id: { Int32: 0, Valid: false },
              support_mode: 'native',
              is_default_route: false,
              is_valid: true,
              is_active: true,
            }, {
              compatibility_id: 32,
              workflow_version_id: 22,
              workflow_version_name: 'Version 2',
              workflow_version_number: 2,
              compatibility_type: 'native',
              workflow_input_mapper_id: { Int32: 0, Valid: false },
              support_mode: 'native',
              is_default_route: true,
              is_valid: true,
              is_active: true,
            }],
            active_tests: [{ id: 41, name: 'load-test-version-split', experiment_type: 'experiment', status: 'active' }],
          }],
        }
      },
    }))

    const { fetchRoutingVersionRows } = useWorkflowRouting()
    const rows = await fetchRoutingVersionRows(7)

    expect(calls).toEqual(['/workflows/7/routing/versions'])
    expect(rows).toEqual([{
      input_schema_id: 11,
      workflow_id: 7,
      input_schema_code: 'v2',
      input_schema_version_number: 2,
      input_schema_status: 'active',
      is_default: true,
      supported_versions: [{
        compatibility_id: 31,
        workflow_version_id: 21,
        workflow_version_name: 'Version 1',
        workflow_version_number: 1,
        compatibility_type: 'native',
        workflow_input_mapper_id: undefined,
        support_mode: 'native',
        is_default_route: false,
        is_valid: true,
        is_active: true,
      }, {
        compatibility_id: 32,
        workflow_version_id: 22,
        workflow_version_name: 'Version 2',
        workflow_version_number: 2,
        compatibility_type: 'native',
        workflow_input_mapper_id: undefined,
        support_mode: 'native',
        is_default_route: true,
        is_valid: true,
        is_active: true,
      }],
      active_tests: [{ id: 41, name: 'load-test-version-split', experiment_type: 'experiment', status: 'active' }],
    }])
  })

  it('fetches routing rows with pagination params and metadata', async () => {
    const calls: Array<[string, unknown]> = []
    vi.stubGlobal('useApi', () => ({
      api: async (url: string, options?: unknown) => {
        calls.push([url, options])
        return {
          success: true,
          data: [],
          meta: { total: 25, page: 2, per_page: 10, total_pages: 3 } satisfies PaginationMeta,
        }
      },
    }))

    const { fetchRoutingVersionRows } = useWorkflowRouting()
    const result = await fetchRoutingVersionRows(7, 2, 10)

    expect(calls).toEqual([[
      '/workflows/7/routing/versions',
      { params: { page: 2, per_page: 10 } },
    ]])
    expect(result).toEqual({
      data: [],
      meta: { total: 25, page: 2, per_page: 10, total_pages: 3 },
    })
  })

  it('treats mapped routing support as mapper even when the raw type is native', async () => {
    vi.stubGlobal('useApi', () => ({
      api: async () => ({
        success: true,
        data: [{
          input_schema_id: 11,
          workflow_id: 7,
          input_schema_code: 'v2',
          input_schema_version_number: 2,
          input_schema_status: 'active',
          is_default: false,
          supported_versions: [{
            compatibility_id: 31,
            workflow_version_id: 21,
            workflow_version_name: 'Version 1',
            workflow_version_number: 1,
            compatibility_type: 'native',
            workflow_input_mapper_id: { Int32: 77, Valid: true },
            support_mode: 'native',
            is_default_route: false,
            is_valid: true,
            is_active: true,
          }],
          active_tests: [],
        }],
      }),
    }))

    const { fetchRoutingVersionRows } = useWorkflowRouting()
    const rows = await fetchRoutingVersionRows(7)

    expect(rows[0].supported_versions[0].workflow_input_mapper_id).toBe(77)
    expect(rows[0].supported_versions[0].support_mode).toBe('mapper')
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

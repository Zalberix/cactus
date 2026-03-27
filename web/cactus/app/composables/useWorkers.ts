import type { ApiResponse } from '~/utils/api-types'

export interface WorkType {
  id: number
  name: string
  code: string
  description?: string
  bootstrap_token?: string
}

export interface Worker {
  id: number
  work_type_id: number
  name: string
  schema_id: number
  status: 'working' | 'ready' | 'offline'
  last_heartbeat_at: string
  work_type_name?: string
}

export interface SettingsRevision {
  id: number
  schema_id: number
  settings_data: Record<string, unknown>
  created_at: string
}

export function useWorkers() {
  const { api } = useApi()

  async function fetchWorkTypes() {
    const resp = await api<ApiResponse<WorkType[]>>('/work-types')
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch work types')
    }
    return resp.data
  }

  async function fetchWorkers(workTypeId: number) {
    const resp = await api<ApiResponse<Worker[]>>(
      `/work-types/${workTypeId}/workers`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch workers')
    }
    return resp.data
  }

  async function fetchWorkersForOrg(_orgId: number): Promise<Worker[]> {
    const workTypes = await fetchWorkTypes()
    const allWorkers: Worker[] = []

    for (const wt of workTypes) {
      try {
        const workers = await fetchWorkers(wt.id)
        for (const w of workers) {
          allWorkers.push({
            ...w,
            work_type_name: wt.name,
          })
        }
      }
      catch {
        // Skip work types with no workers or access errors
      }
    }

    return allWorkers
  }

  async function fetchRevisions(schemaId: number) {
    const resp = await api<ApiResponse<SettingsRevision[]>>(
      `/worker-settings-schemas/${schemaId}/revisions`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch revisions')
    }
    return resp.data
  }

  async function createRevision(schemaId: number, settingsData: Record<string, unknown>) {
    const resp = await api<ApiResponse<SettingsRevision>>(
      `/worker-settings-schemas/${schemaId}/revisions`,
      {
        method: 'POST',
        body: { settings_data: settingsData },
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create revision')
    }
    return resp.data
  }

  return {
    fetchWorkTypes,
    fetchWorkers,
    fetchWorkersForOrg,
    fetchRevisions,
    createRevision,
  }
}

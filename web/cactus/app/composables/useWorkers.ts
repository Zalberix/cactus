import type { ApiResponse } from '~/utils/api-types'

export interface WorkTypeMeta {
  icon?: string
  color?: string
  category?: string
  kind?: string
}

export interface WorkType {
  id: number
  name: string
  code: string
  slug?: string
  description?: string
  meta?: WorkTypeMeta
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

export interface WorkerWorkflowUsage {
  workflow_id: number
  workflow_name: string
  system_id: number
  workflow_version_id: number
  workflow_version_number: number
}

export type DeleteWorkerResult = 'deleted' | 'online' | 'in_use' | 'not_found'

export interface DeleteWorkerResponse {
  result: DeleteWorkerResult
  deleted: boolean
  message: string
  worker_id?: number
  status?: Worker['status']
  usages?: WorkerWorkflowUsage[]
}

export interface SettingsRevision {
  id: number
  schema_id: number
  settings_data: Record<string, unknown>
  created_at: string
}

export interface WorkerSettingsSchema {
  id: number
  work_type_id: number
  version: string
  settings_schema: Record<string, unknown>
  input_schema: Record<string, unknown>
  output_schema: Record<string, unknown>
  created_at: string
}

export interface WorkerSettingsSchemaSummary {
  id: number
  version: string
  created_at: string
  settings_schema: Record<string, unknown>
  worker_count: number
  ready_workers: number
}

export interface WorkTypeCatalogItem extends WorkType {
  worker_count: number
  ready_workers: number
  schemas: WorkerSettingsSchemaSummary[]
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

  async function fetchWorkTypeCatalog(): Promise<WorkTypeCatalogItem[]> {
    const resp = await api<ApiResponse<WorkTypeCatalogItem[]>>('/work-types/catalog')
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch work type catalog')
    }
    return Array.isArray(resp.data) ? resp.data : []
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

  async function deleteWorker(workerId: number): Promise<DeleteWorkerResponse> {
    const resp = await api<ApiResponse<DeleteWorkerResponse>>(
      `/workers/${workerId}`,
      { method: 'DELETE' },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to delete worker')
    }
    return resp.data
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

  async function fetchSchema(schemaId: number): Promise<WorkerSettingsSchema> {
    const resp = await api<ApiResponse<WorkerSettingsSchema>>(
      `/worker-settings-schemas/${schemaId}`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch settings schema')
    }
    return resp.data
  }

  return {
    fetchWorkTypes,
    fetchWorkTypeCatalog,
    fetchWorkers,
    fetchWorkersForOrg,
    deleteWorker,
    fetchRevisions,
    createRevision,
    fetchSchema,
  }
}

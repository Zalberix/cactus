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

export interface WorkerBootstrapToken {
  id: number
  organization_id: number
  work_type_id: number
  name: string
  description?: string
  status: 'active' | 'revoked'
  expires_at?: string
  max_active_workers: number
  active_worker_count: number
  total_registration_count: number
  last_used_at?: string
  created_at: string
  revoked_at?: string
}

export interface CreateWorkerBootstrapTokenInput {
  name: string
  work_type_id: number
  description?: string
  expires_at?: string
  max_active_workers: number
}

export interface CreateWorkerBootstrapTokenResponse {
  token: WorkerBootstrapToken
  plaintext: string
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

  async function fetchWorkerBootstrapTokens(orgId: number): Promise<WorkerBootstrapToken[]> {
    const resp = await api<ApiResponse<WorkerBootstrapToken[]>>(
      `/organizations/${orgId}/worker-bootstrap-tokens`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch worker bootstrap tokens')
    }
    return resp.data
  }

  async function createWorkerBootstrapToken(
    orgId: number,
    input: CreateWorkerBootstrapTokenInput,
  ): Promise<CreateWorkerBootstrapTokenResponse> {
    const resp = await api<ApiResponse<CreateWorkerBootstrapTokenResponse>>(
      `/organizations/${orgId}/worker-bootstrap-tokens`,
      { method: 'POST', body: input },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create worker bootstrap token')
    }
    return resp.data
  }

  async function revokeWorkerBootstrapToken(
    tokenId: number,
    revokeActiveSessions = true,
  ): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/worker-bootstrap-tokens/${tokenId}/revoke`,
      {
        method: 'POST',
        body: { revoke_active_sessions: revokeActiveSessions },
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to revoke worker bootstrap token')
    }
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
    fetchWorkerBootstrapTokens,
    createWorkerBootstrapToken,
    revokeWorkerBootstrapToken,
    deleteWorker,
    fetchRevisions,
    createRevision,
    fetchSchema,
  }
}

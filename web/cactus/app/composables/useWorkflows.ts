import type { ApiResponse } from '~/utils/api-types'

export interface Workflow {
  id: number
  name: string
  priority: number
  system_id: number
  system_name?: string
  version_count?: number
  created_at: string
  updated_at: string
  active_version_count?: number
}

export interface CreateWorkflowRequest {
  name: string
  priority: number
}

export interface WorkflowInputSchemaField {
  name: string
  type: 'string' | 'number' | 'integer' | 'boolean' | 'object' | 'array'
  required: boolean
  description?: string
}

export function workflowOverviewPath(orgId: number, workflowId: number) {
  return `/org/${orgId}/workflows/${workflowId}`
}

export function workflowVersionEditorPath(orgId: number, workflowId: number, versionId: number) {
  return `/org/${orgId}/workflows/${workflowId}/versions/${versionId}/edit`
}

export function workflowVersionCount(workflow: Pick<Workflow, 'version_count' | 'active_version_count'>): number {
  return workflow.version_count ?? workflow.active_version_count ?? 0
}

export function useWorkflows() {
  const { api } = useApi()

  async function fetchWorkflowsForOrg(orgId: number): Promise<Workflow[]> {
    // First get all systems for the org
    const systemsResp = await api<ApiResponse<Array<{ id: number; name: string }>>>(
      `/organizations/${orgId}/systems`,
    )
    if (!systemsResp.success || !systemsResp.data) {
      throw new Error(systemsResp.error?.message ?? 'Failed to fetch systems')
    }

    // Then fetch workflows for each system and aggregate
    const allWorkflows: Workflow[] = []

    await Promise.all(
      systemsResp.data.map(async (system) => {
        const resp = await api<ApiResponse<Workflow[]>>(
          `/systems/${system.id}/workflows`,
        )
        if (resp.success && resp.data) {
          for (const wf of resp.data) {
            allWorkflows.push({
              ...wf,
              system_name: system.name,
            })
          }
        }
      }),
    )

    return allWorkflows
  }

  async function fetchWorkflow(workflowId: number): Promise<Workflow> {
    const resp = await api<ApiResponse<Workflow>>(
      `/workflows/${workflowId}`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch workflow')
    }
    return resp.data
  }

  async function fetchWorkflowsForSystem(systemId: number): Promise<Workflow[]> {
    const resp = await api<ApiResponse<Workflow[]>>(
      `/systems/${systemId}/workflows`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch workflows')
    }
    return Array.isArray(resp.data) ? resp.data : []
  }

  async function createWorkflow(systemId: number, data: CreateWorkflowRequest): Promise<Workflow> {
    const resp = await api<ApiResponse<Workflow>>(
      `/systems/${systemId}/workflows`,
      {
        method: 'POST',
        body: data,
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create workflow')
    }
    return resp.data
  }

  async function updateWorkflow(workflowId: number, data: Partial<CreateWorkflowRequest>): Promise<Workflow | undefined> {
    const resp = await api<ApiResponse<Workflow>>(
      `/workflows/${workflowId}`,
      {
        method: 'PUT',
        body: data,
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update workflow')
    }
    return resp.data
  }

  async function deleteWorkflow(workflowId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/workflows/${workflowId}`,
      {
        method: 'DELETE',
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete workflow')
    }
  }

  return {
    fetchWorkflowsForOrg,
    fetchWorkflowsForSystem,
    fetchWorkflow,
    createWorkflow,
    updateWorkflow,
    deleteWorkflow,
  }
}

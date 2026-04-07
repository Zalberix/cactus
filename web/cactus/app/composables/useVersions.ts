import type { ApiResponse } from '~/utils/api-types'
import type { WorkType, WorkTypeMeta } from '~/composables/useWorkers'

export interface Version {
  id: number
  workflow_id: number
  version_number: number
  is_valid: boolean
  is_active: boolean
  created_at: string
}

export interface Step {
  id: number
  version_id: number
  name: string
  step_type: string
  work_type_id?: number
  work_type_name?: string
  work_type_code?: string
  work_type_meta?: WorkTypeMeta
  control_kind?: string
  control_settings?: Record<string, unknown>
  config?: Record<string, unknown>
  input_mapping?: Record<string, string>
  canvas_position?: { x: number; y: number }
  settings_schema?: Record<string, unknown>
  input_schema?: Record<string, unknown>
  output_schema?: Record<string, unknown>
}

export interface Dependency {
  step_id: number
  depends_on_step_id: number
  outcome: string
  output_index: number
}

export interface ValidationResult {
  valid: boolean
  errors?: string[]
}

export function useVersions() {
  const { api } = useApi()

  async function fetchVersions(workflowId: number): Promise<Version[]> {
    const resp = await api<ApiResponse<Version[]>>(
      `/workflows/${workflowId}/versions`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch versions')
    }
    return resp.data
  }

  async function createVersion(workflowId: number): Promise<Version> {
    const resp = await api<ApiResponse<Version>>(
      `/workflows/${workflowId}/versions`,
      { method: 'POST' },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create version')
    }
    return resp.data
  }

  async function validateVersion(versionId: number): Promise<ValidationResult> {
    const resp = await api<ApiResponse<ValidationResult>>(
      `/versions/${versionId}/validate`,
      { method: 'POST' },
    )
    if (!resp.success || !resp.data) {
      if (resp.error?.details) {
        return {
          valid: false,
          errors: resp.error.details.map(d => d.message),
        }
      }
      throw new Error(resp.error?.message ?? 'Failed to validate version')
    }
    return resp.data
  }

  async function activateVersion(versionId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/versions/${versionId}/activate`,
      { method: 'PATCH' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to activate version')
    }
  }

  async function deactivateVersion(versionId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/versions/${versionId}/deactivate`,
      { method: 'PATCH' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to deactivate version')
    }
  }

  async function fetchSteps(versionId: number): Promise<{ steps: Step[]; dependencies: Dependency[] }> {
    const resp = await api<ApiResponse<{ steps: Step[]; dependencies: Dependency[] }>>(
      `/versions/${versionId}/steps`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch steps')
    }
    return resp.data
  }

  async function createStep(
    versionId: number,
    data: {
      name: string
      step_type: string
      work_type_id?: number
      control_kind?: string
      config?: Record<string, unknown>
      canvas_position?: { x: number; y: number }
    },
  ): Promise<Step> {
    const resp = await api<ApiResponse<Step>>(
      `/versions/${versionId}/steps`,
      {
        method: 'POST',
        body: data,
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create step')
    }
    return resp.data
  }

  async function updateStep(stepId: number, data: Partial<Step>): Promise<Step | undefined> {
    const resp = await api<ApiResponse<Step>>(
      `/steps/${stepId}`,
      {
        method: 'PUT',
        body: data,
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update step')
    }
    return resp.data
  }

  async function updateStepPosition(stepId: number, position: { x: number; y: number }): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}/position`,
      {
        method: 'PATCH',
        body: { canvas_position: position },
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update step position')
    }
  }

  async function deleteStep(stepId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}`,
      { method: 'DELETE' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete step')
    }
  }

  async function createDependency(
    stepId: number,
    dependsOnStepId: number,
    outcome: string,
    outputIndex: number = 0,
  ): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}/dependencies`,
      {
        method: 'POST',
        body: { depends_on_step_id: dependsOnStepId, outcome, output_index: outputIndex },
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to create dependency')
    }
  }

  async function deleteDependency(stepId: number, dependsOnStepId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}/dependencies/${dependsOnStepId}`,
      { method: 'DELETE' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete dependency')
    }
  }

  async function fetchWorkTypes(): Promise<WorkType[]> {
    const resp = await api<ApiResponse<WorkType[]>>('/work-types')
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch work types')
    }
    return resp.data
  }

  return {
    fetchVersions,
    createVersion,
    validateVersion,
    activateVersion,
    deactivateVersion,
    fetchSteps,
    createStep,
    updateStep,
    updateStepPosition,
    deleteStep,
    createDependency,
    deleteDependency,
    fetchWorkTypes,
  }
}

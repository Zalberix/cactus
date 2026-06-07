import type { ApiResponse } from '~/utils/api-types'
import type { WorkType, WorkTypeMeta } from '~/composables/useWorkers'

export interface Version {
  id: number
  workflow_id: number
  name?: string
  version_number: number
  is_valid: boolean
  is_active: boolean
  created_at: string
}

export interface VersionSummary {
  id: number
  workflow_id: number
  name: string
  version_number: number
  is_valid: boolean
  is_active: boolean
  run_count: number
  created_at: string
  updated_at?: string
  deleted_at?: string
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
  input_mapping?: Record<string, string> | Array<{ target: string, source: string }>
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

export interface ValidationIssue {
  type?: string
  step_id?: number
  field?: string
  message: string
}

export interface ValidationResult {
  isValid: boolean
  errors: ValidationIssue[]
}

export function normalizeValidationIssue(value: unknown): ValidationIssue | null {
  if (typeof value === 'string' && value) return { message: value }
  if (!value || typeof value !== 'object') return null

  const raw = value as Record<string, unknown>
  const message = typeof raw.message === 'string' ? raw.message : ''
  if (!message) return null

  const issue: ValidationIssue = { message }
  if (typeof raw.type === 'string') issue.type = raw.type
  if (typeof raw.field === 'string') issue.field = raw.field
  if (typeof raw.step_id === 'number') issue.step_id = raw.step_id
  return issue
}

export function normalizeValidationResult(payload: {
  is_valid?: boolean
  valid?: boolean
  errors?: unknown
} | undefined | null): ValidationResult {
  const rawErrors = Array.isArray(payload?.errors) ? payload.errors : []
  const errors = rawErrors
    .map(normalizeValidationIssue)
    .filter((value): value is ValidationIssue => Boolean(value))

  return {
    isValid: Boolean(payload?.is_valid ?? payload?.valid ?? false),
    errors,
  }
}

function extractValidationErrorMessages(error: unknown): ValidationIssue[] | null {
  if (!error || typeof error !== 'object') return null

  const response = (error as {
    response?: { _data?: ApiResponse<{ is_valid?: boolean; valid?: boolean; errors?: unknown }> }
    data?: ApiResponse<{ is_valid?: boolean; valid?: boolean; errors?: unknown }>
  })
  const responsePayload = response.response?._data ?? response.data
  if (!responsePayload || typeof responsePayload !== 'object') return null

  if (!responsePayload.error || !Array.isArray(responsePayload.error.details)) return null

  const messages = responsePayload.error.details
    .map(normalizeValidationIssue)
    .filter((value): value is ValidationIssue => Boolean(value))

  return messages.length > 0 ? messages : null
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
    return Array.isArray(resp.data) ? resp.data : []
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

  async function fetchVersionSummaries(workflowId: number): Promise<VersionSummary[]> {
    const resp = await api<ApiResponse<VersionSummary[]>>(
      `/workflows/${workflowId}/version-summaries`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch version summaries')
    }
    return Array.isArray(resp.data) ? resp.data : []
  }

  async function copyVersion(versionId: number): Promise<Version> {
    const resp = await api<ApiResponse<Version>>(
      `/versions/${versionId}/copy`,
      { method: 'POST' },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to copy version')
    }
    return resp.data
  }

  async function updateVersionName(versionId: number, name: string): Promise<Version> {
    const resp = await api<ApiResponse<Version>>(
      `/versions/${versionId}/name`,
      {
        method: 'PATCH',
        body: { name },
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to update version name')
    }
    return resp.data
  }

  async function validateVersion(versionId: number): Promise<ValidationResult> {
    try {
      const resp = await api<ApiResponse<{ is_valid?: boolean; valid?: boolean; errors?: unknown }>>(
        `/versions/${versionId}/validate`,
        { method: 'POST' },
      )

      if (resp.success && resp.data) {
        return normalizeValidationResult(resp.data)
      }

      const errorMessages = extractValidationErrorMessages(resp)
      if (errorMessages) {
        return {
          isValid: false,
          errors: errorMessages,
        }
      }

      throw new Error(resp.error?.message ?? 'Failed to validate version')
    }
    catch (error) {
      const errorMessages = extractValidationErrorMessages(error)
      if (errorMessages) {
        return {
          isValid: false,
          errors: errorMessages,
        }
      }

      throw new Error(error instanceof Error ? error.message : 'Failed to validate version')
    }
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

  async function deleteVersion(versionId: number): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/versions/${versionId}`,
      { method: 'DELETE' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete version')
    }
  }

  async function fetchSteps(versionId: number): Promise<{ steps: Step[]; dependencies: Dependency[] }> {
    const resp = await api<ApiResponse<{ steps: Step[]; dependencies: Dependency[] }>>(
      `/versions/${versionId}/steps`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch steps')
    }
    return {
      steps: Array.isArray(resp.data.steps) ? resp.data.steps : [],
      dependencies: Array.isArray(resp.data.dependencies) ? resp.data.dependencies : [],
    }
  }

  async function createStep(
    versionId: number,
    data: {
      name: string
      step_type: string
      work_type_id?: number
      worker_settings_schema_id?: number
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

  async function updateTaskSettings(stepId: number, settingsData: Record<string, unknown>): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}/task-settings`,
      {
        method: 'PUT',
        body: { settings_data: settingsData },
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update task settings')
    }
  }

  async function updateTaskInputMapping(stepId: number, inputMapping: Array<{ target: string, source: string }>): Promise<void> {
    const resp = await api<ApiResponse<null>>(
      `/steps/${stepId}/input-mapping`,
      {
        method: 'PUT',
        body: { input_mapping: inputMapping },
      },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update input mapping')
    }
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
    return Array.isArray(resp.data) ? resp.data : []
  }

  return {
    fetchVersions,
    createVersion,
    fetchVersionSummaries,
    copyVersion,
    updateVersionName,
    validateVersion,
    activateVersion,
    deactivateVersion,
    deleteVersion,
    fetchSteps,
    createStep,
    updateStep,
    updateTaskSettings,
    updateTaskInputMapping,
    updateStepPosition,
    deleteStep,
    createDependency,
    deleteDependency,
    fetchWorkTypes,
  }
}

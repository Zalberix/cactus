import type { ApiResponse } from '~/utils/api-types'

export interface WorkflowInputSchemaRecord {
  id: number
  workflow_id: number
  code: string
  version_number: number
  schema_json: Record<string, unknown>
  status: 'draft' | 'active' | 'deprecated' | 'archived'
  is_default: boolean
}

export interface WorkflowInputMapperRecord {
  id: number
  workflow_id: number
  name: string
  mapper_type: string
  rules: Record<string, unknown>
  is_active: boolean
}

export interface WorkflowSchemaCompatibility {
  id: number
  workflow_version_id: number
  workflow_input_schema_id: number
  compatibility_type: string
  workflow_input_mapper_id?: number
  default_values: Record<string, unknown>
  is_active: boolean
  is_default_route: boolean
}

export interface WorkflowExperiment {
  id: number
  workflow_id: number
  name: string
  description?: string
  experiment_type: string
  status: string
  started_at?: string
  ended_at?: string
}

export interface WorkflowExperimentScope {
  id: number
  workflow_experiment_id: number
  workflow_input_schema_id: number
  traffic_conditions: Record<string, unknown>
  conditions_hash: string
  traffic_percent: number
  fallback_policy: string
  fallback_workflow_version_id?: number
}

export interface WorkflowExperimentVariant {
  id: number
  workflow_experiment_scope_id: number
  workflow_version_id: number
  traffic_weight: number
  is_control_group: boolean
  is_active: boolean
}

export interface CreateInputSchemaPayload {
  code: string
  version_number?: number
  schema_json: Record<string, unknown>
  status?: string
  is_default?: boolean
}

export function workflowRoutingPath(orgId: number, workflowId: number) {
  return `/org/${orgId}/workflows/${workflowId}/routing`
}

export function workflowInputSchemaCreatePath(orgId: number, workflowId: number) {
  return `${workflowRoutingPath(orgId, workflowId)}/input-schemas/new`
}

export function workflowInputSchemaEditorPath(orgId: number, workflowId: number, inputSchemaId: number) {
  return `${workflowRoutingPath(orgId, workflowId)}/input-schemas/${inputSchemaId}/edit`
}

export function nativeInputSchemaForVersion(
  versionId: number,
  schemas: WorkflowInputSchemaRecord[],
  compatibilities: WorkflowSchemaCompatibility[],
): WorkflowInputSchemaRecord | undefined {
  const compatibility = compatibilities.find(item =>
    item.workflow_version_id === versionId
    && item.compatibility_type === 'native'
    && item.is_active,
  )
  if (!compatibility) return undefined
  return schemas.find(schema => schema.id === compatibility.workflow_input_schema_id)
}

export interface CreateInputMapperPayload {
  name: string
  mapper_type: string
  rules: Record<string, unknown>
  is_active?: boolean
}

export interface CreateCompatibilityPayload {
  workflow_input_schema_id: number
  compatibility_type: string
  workflow_input_mapper_id?: number
  default_values: Record<string, unknown>
  is_active?: boolean
  is_default_route?: boolean
}

export interface CreateExperimentPayload {
  name: string
  description?: string
  experiment_type?: string
  status?: string
  started_at?: string
  ended_at?: string
}

export interface CreateExperimentScopePayload {
  workflow_input_schema_id: number
  traffic_conditions: Record<string, unknown>
  traffic_percent?: number
  fallback_policy?: string
  fallback_workflow_version_id?: number
}

export interface CreateExperimentVariantPayload {
  workflow_version_id: number
  traffic_weight?: number
  is_control_group?: boolean
  is_active?: boolean
}

function requireData<T>(resp: ApiResponse<T>, fallback: string): T {
  if (!resp.success || resp.data === undefined || resp.data === null) {
    throw new Error(resp.error?.message ?? fallback)
  }
  return resp.data
}

function recordValue(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function nullableNumber(value: unknown): number | undefined {
  if (typeof value === 'number') return value
  const raw = recordValue(value)
  if (Object.keys(raw).length === 0) return undefined
  const valid = raw.Valid ?? raw.valid
  if (valid === false) return undefined
  const candidate = raw.Int32 ?? raw.int32 ?? raw.Int64 ?? raw.int64 ?? raw.Value ?? raw.value
  return typeof candidate === 'number' ? candidate : undefined
}

function nullableString(value: unknown): string | undefined {
  if (typeof value === 'string') return value
  const raw = recordValue(value)
  if (Object.keys(raw).length === 0) return undefined
  const valid = raw.Valid ?? raw.valid
  if (valid === false) return undefined
  const candidate = raw.String ?? raw.string ?? raw.Time ?? raw.time ?? raw.Value ?? raw.value
  return typeof candidate === 'string' ? candidate : undefined
}

function jsonObject(value: unknown): Record<string, unknown> {
  return recordValue(value)
}

function normalizeInputSchema(value: unknown): WorkflowInputSchemaRecord {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_id: Number(raw.workflow_id ?? 0),
    code: String(raw.code ?? ''),
    version_number: Number(raw.version_number ?? 0),
    schema_json: jsonObject(raw.schema_json),
    status: raw.status ?? 'draft',
    is_default: Boolean(raw.is_default),
  }
}

function normalizeInputMapper(value: unknown): WorkflowInputMapperRecord {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_id: Number(raw.workflow_id ?? 0),
    name: String(raw.name ?? ''),
    mapper_type: String(raw.mapper_type ?? ''),
    rules: jsonObject(raw.rules),
    is_active: Boolean(raw.is_active),
  }
}

function normalizeCompatibility(value: unknown): WorkflowSchemaCompatibility {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_version_id: Number(raw.workflow_version_id ?? 0),
    workflow_input_schema_id: Number(raw.workflow_input_schema_id ?? 0),
    compatibility_type: String(raw.compatibility_type ?? ''),
    workflow_input_mapper_id: nullableNumber(raw.workflow_input_mapper_id),
    default_values: jsonObject(raw.default_values),
    is_active: Boolean(raw.is_active),
    is_default_route: Boolean(raw.is_default_route),
  }
}

function normalizeExperiment(value: unknown): WorkflowExperiment {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_id: Number(raw.workflow_id ?? 0),
    name: String(raw.name ?? ''),
    description: nullableString(raw.description),
    experiment_type: String(raw.experiment_type ?? ''),
    status: String(raw.status ?? ''),
    started_at: nullableString(raw.started_at),
    ended_at: nullableString(raw.ended_at),
  }
}

function normalizeExperimentScope(value: unknown): WorkflowExperimentScope {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_experiment_id: Number(raw.workflow_experiment_id ?? 0),
    workflow_input_schema_id: Number(raw.workflow_input_schema_id ?? 0),
    traffic_conditions: jsonObject(raw.traffic_conditions),
    conditions_hash: String(raw.conditions_hash ?? ''),
    traffic_percent: Number(raw.traffic_percent ?? 0),
    fallback_policy: String(raw.fallback_policy ?? ''),
    fallback_workflow_version_id: nullableNumber(raw.fallback_workflow_version_id),
  }
}

function normalizeExperimentVariant(value: unknown): WorkflowExperimentVariant {
  const raw = recordValue(value)
  return {
    id: Number(raw.id ?? 0),
    workflow_experiment_scope_id: Number(raw.workflow_experiment_scope_id ?? 0),
    workflow_version_id: Number(raw.workflow_version_id ?? 0),
    traffic_weight: Number(raw.traffic_weight ?? 0),
    is_control_group: Boolean(raw.is_control_group),
    is_active: Boolean(raw.is_active),
  }
}

export function useWorkflowRouting() {
  const { api } = useApi()

  async function fetchInputSchemas(workflowId: number): Promise<WorkflowInputSchemaRecord[]> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord[]>>(`/workflows/${workflowId}/input-schemas`)
    const data = requireData(resp, 'Failed to fetch input schemas')
    return Array.isArray(data) ? data.map(normalizeInputSchema) : []
  }

  async function fetchInputSchema(inputSchemaId: number): Promise<WorkflowInputSchemaRecord> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord>>(`/input-schemas/${inputSchemaId}`)
    return normalizeInputSchema(requireData(resp, 'Failed to fetch input schema'))
  }

  async function createInputSchema(workflowId: number, payload: CreateInputSchemaPayload): Promise<WorkflowInputSchemaRecord> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord>>(`/workflows/${workflowId}/input-schemas`, {
      method: 'POST',
      body: payload,
    })
    return normalizeInputSchema(requireData(resp, 'Failed to create input schema'))
  }

  async function updateInputSchema(inputSchemaId: number, payload: Partial<CreateInputSchemaPayload>): Promise<WorkflowInputSchemaRecord> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord>>(`/input-schemas/${inputSchemaId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeInputSchema(requireData(resp, 'Failed to update input schema'))
  }

  async function setInputSchemaDefault(inputSchemaId: number): Promise<WorkflowInputSchemaRecord> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord>>(`/input-schemas/${inputSchemaId}/default`, {
      method: 'PATCH',
    })
    return normalizeInputSchema(requireData(resp, 'Failed to set default input schema'))
  }

  async function updateInputSchemaStatus(inputSchemaId: number, status: string): Promise<WorkflowInputSchemaRecord> {
    const resp = await api<ApiResponse<WorkflowInputSchemaRecord>>(`/input-schemas/${inputSchemaId}/status`, {
      method: 'PATCH',
      body: { status },
    })
    return normalizeInputSchema(requireData(resp, 'Failed to update input schema status'))
  }

  async function deleteInputSchema(inputSchemaId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/input-schemas/${inputSchemaId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete input schema')
  }

  async function fetchInputMappers(workflowId: number): Promise<WorkflowInputMapperRecord[]> {
    const resp = await api<ApiResponse<WorkflowInputMapperRecord[]>>(`/workflows/${workflowId}/input-mappers`)
    const data = requireData(resp, 'Failed to fetch input mappers')
    return Array.isArray(data) ? data.map(normalizeInputMapper) : []
  }

  async function createInputMapper(workflowId: number, payload: CreateInputMapperPayload): Promise<WorkflowInputMapperRecord> {
    const resp = await api<ApiResponse<WorkflowInputMapperRecord>>(`/workflows/${workflowId}/input-mappers`, {
      method: 'POST',
      body: payload,
    })
    return normalizeInputMapper(requireData(resp, 'Failed to create input mapper'))
  }

  async function updateInputMapper(inputMapperId: number, payload: Partial<CreateInputMapperPayload>): Promise<WorkflowInputMapperRecord> {
    const resp = await api<ApiResponse<WorkflowInputMapperRecord>>(`/input-mappers/${inputMapperId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeInputMapper(requireData(resp, 'Failed to update input mapper'))
  }

  async function setInputMapperActive(inputMapperId: number, isActive: boolean): Promise<WorkflowInputMapperRecord> {
    const resp = await api<ApiResponse<WorkflowInputMapperRecord>>(`/input-mappers/${inputMapperId}/active`, {
      method: 'PATCH',
      body: { is_active: isActive },
    })
    return normalizeInputMapper(requireData(resp, 'Failed to update input mapper active state'))
  }

  async function deleteInputMapper(inputMapperId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/input-mappers/${inputMapperId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete input mapper')
  }

  async function fetchCompatibilities(workflowId: number): Promise<WorkflowSchemaCompatibility[]> {
    const resp = await api<ApiResponse<WorkflowSchemaCompatibility[]>>(`/workflows/${workflowId}/schema-compatibilities`)
    const data = requireData(resp, 'Failed to fetch schema compatibilities')
    return Array.isArray(data) ? data.map(normalizeCompatibility) : []
  }

  async function createCompatibility(versionId: number, payload: CreateCompatibilityPayload): Promise<WorkflowSchemaCompatibility> {
    const resp = await api<ApiResponse<WorkflowSchemaCompatibility>>(`/versions/${versionId}/schema-compatibilities`, {
      method: 'POST',
      body: payload,
    })
    return normalizeCompatibility(requireData(resp, 'Failed to create schema compatibility'))
  }

  async function updateCompatibility(compatibilityId: number, payload: Partial<CreateCompatibilityPayload>): Promise<WorkflowSchemaCompatibility> {
    const resp = await api<ApiResponse<WorkflowSchemaCompatibility>>(`/schema-compatibilities/${compatibilityId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeCompatibility(requireData(resp, 'Failed to update schema compatibility'))
  }

  async function setCompatibilityDefaultRoute(compatibilityId: number, isDefaultRoute = true): Promise<WorkflowSchemaCompatibility> {
    const resp = await api<ApiResponse<WorkflowSchemaCompatibility>>(`/schema-compatibilities/${compatibilityId}/default-route`, {
      method: 'PATCH',
      body: { is_default_route: isDefaultRoute },
    })
    return normalizeCompatibility(requireData(resp, 'Failed to update default route'))
  }

  async function deactivateCompatibility(compatibilityId: number): Promise<WorkflowSchemaCompatibility> {
    const resp = await api<ApiResponse<WorkflowSchemaCompatibility>>(`/schema-compatibilities/${compatibilityId}/deactivate`, {
      method: 'PATCH',
    })
    return normalizeCompatibility(requireData(resp, 'Failed to deactivate schema compatibility'))
  }

  async function deleteCompatibility(compatibilityId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/schema-compatibilities/${compatibilityId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete schema compatibility')
  }

  async function fetchExperiments(workflowId: number): Promise<WorkflowExperiment[]> {
    const resp = await api<ApiResponse<WorkflowExperiment[]>>(`/workflows/${workflowId}/experiments`)
    const data = requireData(resp, 'Failed to fetch experiments')
    return Array.isArray(data) ? data.map(normalizeExperiment) : []
  }

  async function createExperiment(workflowId: number, payload: CreateExperimentPayload): Promise<WorkflowExperiment> {
    const resp = await api<ApiResponse<WorkflowExperiment>>(`/workflows/${workflowId}/experiments`, {
      method: 'POST',
      body: payload,
    })
    return normalizeExperiment(requireData(resp, 'Failed to create experiment'))
  }

  async function updateExperiment(experimentId: number, payload: Partial<CreateExperimentPayload>): Promise<WorkflowExperiment> {
    const resp = await api<ApiResponse<WorkflowExperiment>>(`/experiments/${experimentId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeExperiment(requireData(resp, 'Failed to update experiment'))
  }

  async function updateExperimentStatus(experimentId: number, status: string): Promise<WorkflowExperiment> {
    const resp = await api<ApiResponse<WorkflowExperiment>>(`/experiments/${experimentId}/status`, {
      method: 'PATCH',
      body: { status },
    })
    return normalizeExperiment(requireData(resp, 'Failed to update experiment status'))
  }

  async function deleteExperiment(experimentId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/experiments/${experimentId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete experiment')
  }

  async function fetchExperimentScopes(experimentId: number): Promise<WorkflowExperimentScope[]> {
    const resp = await api<ApiResponse<WorkflowExperimentScope[]>>(`/experiments/${experimentId}/scopes`)
    const data = requireData(resp, 'Failed to fetch experiment scopes')
    return Array.isArray(data) ? data.map(normalizeExperimentScope) : []
  }

  async function createExperimentScope(experimentId: number, payload: CreateExperimentScopePayload): Promise<WorkflowExperimentScope> {
    const resp = await api<ApiResponse<WorkflowExperimentScope>>(`/experiments/${experimentId}/scopes`, {
      method: 'POST',
      body: payload,
    })
    return normalizeExperimentScope(requireData(resp, 'Failed to create experiment scope'))
  }

  async function updateExperimentScope(scopeId: number, payload: Partial<CreateExperimentScopePayload>): Promise<WorkflowExperimentScope> {
    const resp = await api<ApiResponse<WorkflowExperimentScope>>(`/experiment-scopes/${scopeId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeExperimentScope(requireData(resp, 'Failed to update experiment scope'))
  }

  async function deleteExperimentScope(scopeId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/experiment-scopes/${scopeId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete experiment scope')
  }

  async function fetchExperimentVariants(scopeId: number): Promise<WorkflowExperimentVariant[]> {
    const resp = await api<ApiResponse<WorkflowExperimentVariant[]>>(`/experiment-scopes/${scopeId}/variants`)
    const data = requireData(resp, 'Failed to fetch experiment variants')
    return Array.isArray(data) ? data.map(normalizeExperimentVariant) : []
  }

  async function createExperimentVariant(scopeId: number, payload: CreateExperimentVariantPayload): Promise<WorkflowExperimentVariant> {
    const resp = await api<ApiResponse<WorkflowExperimentVariant>>(`/experiment-scopes/${scopeId}/variants`, {
      method: 'POST',
      body: payload,
    })
    return normalizeExperimentVariant(requireData(resp, 'Failed to create experiment variant'))
  }

  async function updateExperimentVariant(variantId: number, payload: Partial<CreateExperimentVariantPayload>): Promise<WorkflowExperimentVariant> {
    const resp = await api<ApiResponse<WorkflowExperimentVariant>>(`/experiment-variants/${variantId}`, {
      method: 'PUT',
      body: payload,
    })
    return normalizeExperimentVariant(requireData(resp, 'Failed to update experiment variant'))
  }

  async function deleteExperimentVariant(variantId: number): Promise<void> {
    const resp = await api<ApiResponse<unknown>>(`/experiment-variants/${variantId}`, {
      method: 'DELETE',
    })
    if (!resp.success) throw new Error(resp.error?.message ?? 'Failed to delete experiment variant')
  }

  return {
    fetchInputSchemas,
    fetchInputSchema,
    createInputSchema,
    updateInputSchema,
    setInputSchemaDefault,
    updateInputSchemaStatus,
    deleteInputSchema,
    fetchInputMappers,
    createInputMapper,
    updateInputMapper,
    setInputMapperActive,
    deleteInputMapper,
    fetchCompatibilities,
    createCompatibility,
    updateCompatibility,
    setCompatibilityDefaultRoute,
    deactivateCompatibility,
    deleteCompatibility,
    fetchExperiments,
    createExperiment,
    updateExperiment,
    updateExperimentStatus,
    deleteExperiment,
    fetchExperimentScopes,
    createExperimentScope,
    updateExperimentScope,
    deleteExperimentScope,
    fetchExperimentVariants,
    createExperimentVariant,
    updateExperimentVariant,
    deleteExperimentVariant,
  }
}

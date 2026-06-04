import type {
  WorkflowInputSchemaRecord,
  WorkflowSchemaCompatibility,
} from '~/composables/useWorkflowRouting'

type SchemaLabelSource = Pick<WorkflowInputSchemaRecord, 'id' | 'code' | 'version_number'>
type CompatibilitySource = Pick<
  WorkflowSchemaCompatibility,
  'workflow_version_id' | 'workflow_input_schema_id' | 'is_active'
>

export type SupportedSchemaLabelsByVersionId = Record<number, string[]>

export function schemaVersionLabel(schema: Pick<WorkflowInputSchemaRecord, 'code' | 'version_number'>) {
  return `${schema.code} v${schema.version_number}`
}

export function buildSupportedSchemaLabelsByVersionId(
  schemas: SchemaLabelSource[],
  compatibilities: CompatibilitySource[],
): SupportedSchemaLabelsByVersionId {
  const schemaById = new Map(schemas.map(schema => [schema.id, schema]))
  const labelsByVersionId: SupportedSchemaLabelsByVersionId = {}

  for (const compatibility of compatibilities) {
    if (!compatibility.is_active) continue

    const schema = schemaById.get(compatibility.workflow_input_schema_id)
    if (!schema) continue

    const labels = labelsByVersionId[compatibility.workflow_version_id] ?? []
    const label = schemaVersionLabel(schema)
    if (!labels.includes(label)) {
      labels.push(label)
      labelsByVersionId[compatibility.workflow_version_id] = labels
    }
  }

  for (const labels of Object.values(labelsByVersionId)) {
    labels.sort((left, right) => left.localeCompare(right, undefined, {
      numeric: true,
      sensitivity: 'base',
    }))
  }

  return labelsByVersionId
}

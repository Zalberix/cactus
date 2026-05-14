export function workflowInputPath(name: string) {
  return `$.message.value.${name}`
}

export interface WorkflowInputField {
  name: string
  type: 'string' | 'number' | 'integer' | 'boolean' | 'object' | 'array'
  required: boolean
  description?: string
}

const knownTypes = new Set<WorkflowInputField['type']>([
  'string',
  'number',
  'integer',
  'boolean',
  'object',
  'array',
])

function schemaType(value: unknown): WorkflowInputField['type'] {
  return typeof value === 'string' && knownTypes.has(value as WorkflowInputField['type'])
    ? value as WorkflowInputField['type']
    : 'string'
}

export function workflowInputFieldsFromSchema(schema: Record<string, unknown> | null | undefined): WorkflowInputField[] {
  const properties = schema?.properties
  if (!properties || typeof properties !== 'object' || Array.isArray(properties)) return []

  return Object.entries(properties as Record<string, unknown>)
    .map(([name, property]) => {
      const prop = property && typeof property === 'object' && !Array.isArray(property)
        ? property as Record<string, unknown>
        : {}

      const field: WorkflowInputField = {
        name,
        type: schemaType(prop.type),
        required: prop.required === true,
      }
      if (typeof prop.description === 'string' && prop.description) {
        field.description = prop.description
      }
      return field
    })
    .sort((a, b) => a.name.localeCompare(b.name))
}

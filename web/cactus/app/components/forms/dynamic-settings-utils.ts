export interface JsonSchemaProperty {
  type?: string
  title?: string
  description?: string
  default?: unknown
  enum?: string[]
  required?: boolean
  'x-ui-widget'?: string
  'x-ui-order'?: number
  'x-ui-placeholder'?: string
  'x-ui-help'?: string
}

export interface JsonSchema {
  type?: string
  required?: string[]
  properties?: Record<string, JsonSchemaProperty>
}

export function schemaPropertiesForForm(schema: JsonSchema) {
  const required = new Set(schema.required ?? [])
  return Object.entries(schema.properties ?? {})
    .map(([key, prop]) => ({
      key,
      ...prop,
      isRequired: required.has(key) || prop.required === true,
    }))
    .sort((a, b) => (a['x-ui-order'] ?? 999) - (b['x-ui-order'] ?? 999))
}

export function validateSettingsData(schema: JsonSchema, data: Record<string, unknown>): string[] {
  const errors: string[] = []
  for (const prop of schemaPropertiesForForm(schema)) {
    const value = data[prop.key]
    if (isEmptySetting(value)) {
      if (prop.isRequired) errors.push(prop.key)
      continue
    }
    if (prop.type && !matchesSchemaType(prop.type, value)) {
      errors.push(prop.key)
      continue
    }
    if (prop.enum?.length && typeof value === 'string' && !prop.enum.includes(value)) {
      errors.push(prop.key)
    }
  }
  return [...new Set(errors)]
}

function isEmptySetting(value: unknown): boolean {
  return value === undefined || value === null || value === ''
}

function matchesSchemaType(type: string, value: unknown): boolean {
  switch (type) {
    case 'string':
      return typeof value === 'string'
    case 'number':
      return typeof value === 'number' && Number.isFinite(value)
    case 'integer':
      return typeof value === 'number' && Number.isInteger(value)
    case 'boolean':
      return typeof value === 'boolean'
    case 'object':
      return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
    case 'array':
      return Array.isArray(value)
    default:
      return true
  }
}

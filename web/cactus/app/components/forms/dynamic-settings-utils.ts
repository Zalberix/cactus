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

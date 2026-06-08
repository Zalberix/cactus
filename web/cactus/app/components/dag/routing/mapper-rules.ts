export type MapperValueType = 'string' | 'number' | 'boolean' | 'null' | 'json'

export interface MapperDefaultRow {
  id: string
  key: string
  value: string
  valueType: MapperValueType
}

export interface MapperMappingRow {
  id: string
  targetPath: string
  mode: 'source' | 'literal'
  sourcePath: string
  literalValue: string
  literalType: MapperValueType
}

export interface MapperRulesBuilderState {
  copyAll: boolean
  defaults: MapperDefaultRow[]
  mappings: MapperMappingRow[]
}

let rowCounter = 0

export function createMapperDefaultRow(key = '', value: unknown = ''): MapperDefaultRow {
  const typed = typedValueFromUnknown(value)
  return {
    id: nextRowID('default'),
    key,
    value: typed.value,
    valueType: typed.valueType,
  }
}

export function createMapperMappingRow(targetPath = '', sourcePath = ''): MapperMappingRow {
  return {
    id: nextRowID('mapping'),
    targetPath,
    mode: sourcePath ? 'source' : 'literal',
    sourcePath,
    literalValue: '',
    literalType: 'string',
  }
}

export function rulesToBuilderState(rawRules: string): MapperRulesBuilderState {
  const rules = parseRules(rawRules)
  const mappings = objectValue(rules.mapping) ?? objectValue(rules.fields) ?? objectValue(rules.mappings) ?? {}
  return {
    copyAll: rules.copy_all === true,
    defaults: Object.entries(objectValue(rules.defaults) ?? {}).map(([key, value]) => createMapperDefaultRow(key, value)),
    mappings: Object.entries(mappings).map(([targetPath, sourceSpec]) => mappingRowFromSourceSpec(targetPath, sourceSpec)),
  }
}

export function builderStateToRules(state: MapperRulesBuilderState): Record<string, unknown> {
  const defaults: Record<string, unknown> = {}
  for (const row of state.defaults) {
    const key = row.key.trim()
    if (!key) continue
    defaults[key] = parseTypedValue(row.value, row.valueType)
  }

  const mapping: Record<string, unknown> = {}
  for (const row of state.mappings) {
    const targetPath = row.targetPath.trim()
    if (!targetPath) continue

    if (row.mode === 'source') {
      const sourcePath = row.sourcePath.trim()
      if (!sourcePath) continue
      mapping[targetPath] = sourcePath
    }
    else {
      mapping[targetPath] = {
        literal: parseTypedValue(row.literalValue, row.literalType),
      }
    }
  }

  const rules: Record<string, unknown> = {
    copy_all: state.copyAll,
  }
  if (Object.keys(defaults).length > 0) {
    rules.defaults = defaults
  }
  if (Object.keys(mapping).length > 0) {
    rules.mapping = mapping
  }
  return rules
}

export function builderStateToRulesJson(state: MapperRulesBuilderState): string {
  return JSON.stringify(builderStateToRules(state), null, 2)
}

export function mapperBuilderWarnings(state: MapperRulesBuilderState): string[] {
  const warnings: string[] = []
  const targets = new Set<string>()
  const duplicateTargets = new Set<string>()

  for (const row of state.mappings) {
    const targetPath = row.targetPath.trim()
    if (!targetPath) continue
    if (targets.has(targetPath)) {
      duplicateTargets.add(targetPath)
    }
    targets.add(targetPath)
  }

  if (duplicateTargets.size > 0) {
    warnings.push(`Duplicate target paths: ${Array.from(duplicateTargets).join(', ')}`)
  }
  return warnings
}

function mappingRowFromSourceSpec(targetPath: string, sourceSpec: unknown): MapperMappingRow {
  const row = createMapperMappingRow(targetPath)
  if (typeof sourceSpec === 'string') {
    row.mode = 'source'
    row.sourcePath = sourceSpec
    return row
  }

  const objectSpec = objectValue(sourceSpec)
  if (!objectSpec) {
    row.mode = 'literal'
    return row
  }

  if ('literal' in objectSpec) {
    const typed = typedValueFromUnknown(objectSpec.literal)
    row.mode = 'literal'
    row.literalValue = typed.value
    row.literalType = typed.valueType
    return row
  }

  const sourcePath = stringValue(objectSpec.source) ?? stringValue(objectSpec.path) ?? ''
  row.mode = sourcePath ? 'source' : 'literal'
  row.sourcePath = sourcePath
  return row
}

function parseRules(rawRules: string): Record<string, unknown> {
  try {
    const value = JSON.parse(rawRules || '{}')
    return objectValue(value) ?? {}
  }
  catch {
    return {}
  }
}

function parseTypedValue(value: string, valueType: MapperValueType): unknown {
  switch (valueType) {
    case 'number': {
      const parsed = Number(value)
      return Number.isFinite(parsed) ? parsed : 0
    }
    case 'boolean':
      return value === 'true'
    case 'null':
      return null
    case 'json':
      try {
        return JSON.parse(value)
      }
      catch {
        return value
      }
    case 'string':
    default:
      return value
  }
}

function typedValueFromUnknown(value: unknown): { value: string, valueType: MapperValueType } {
  if (value === null) {
    return { value: '', valueType: 'null' }
  }
  if (typeof value === 'number') {
    return { value: String(value), valueType: 'number' }
  }
  if (typeof value === 'boolean') {
    return { value: String(value), valueType: 'boolean' }
  }
  if (typeof value === 'string') {
    return { value, valueType: 'string' }
  }
  return { value: JSON.stringify(value, null, 2), valueType: 'json' }
}

function objectValue(value: unknown): Record<string, unknown> | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  return value as Record<string, unknown>
}

function stringValue(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined
}

function nextRowID(prefix: string): string {
  rowCounter += 1
  return `${prefix}-${rowCounter}`
}

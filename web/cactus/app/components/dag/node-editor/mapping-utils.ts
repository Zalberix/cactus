export function mappingRecordToEntries(mapping: Record<string, string>) {
  return Object.entries(mapping)
    .filter(([target, source]) => target.trim() && source.trim())
    .map(([target, source]) => ({ target, source }))
}

export function setMappingExpression(
  mapping: Record<string, string>,
  target: string | null,
  expression: string,
) {
  if (!target) return mapping
  return {
    ...mapping,
    [target]: expression,
  }
}

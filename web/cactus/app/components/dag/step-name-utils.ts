const MAX_STEP_NAME_LENGTH = 255

export interface StepNameEntry {
  id: string
  name: string
}

export function normalizeStepName(name: string): string {
  return name.trim()
}

export function isStepNameDuplicate(
  name: string,
  names: StepNameEntry[],
  excludeId?: string,
): boolean {
  const normalized = normalizeStepName(name)
  if (!normalized) return false

  return names.some(item =>
    item.id !== excludeId && normalizeStepName(item.name) === normalized,
  )
}

export function nextUniqueStepName(baseName: string, existingNames: string[]): string {
  const base = truncateStepName(normalizeStepName(baseName) || 'Step')
  const used = new Set(
    existingNames
      .map(normalizeStepName)
      .filter(Boolean),
  )

  if (!used.has(base)) return base

  let maxSuffix = 1
  const prefix = `${base} `
  for (const name of used) {
    if (!name.startsWith(prefix)) continue
    const suffix = Number(name.slice(prefix.length))
    if (Number.isInteger(suffix) && suffix >= 2) {
      maxSuffix = Math.max(maxSuffix, suffix)
    }
  }

  for (let suffix = maxSuffix + 1; ; suffix++) {
    const suffixText = ` ${suffix}`
    const candidate = `${truncateStepNameForSuffix(base, suffixText)}${suffixText}`
    if (!used.has(candidate)) return candidate
  }
}

function truncateStepName(name: string): string {
  return [...name].slice(0, MAX_STEP_NAME_LENGTH).join('')
}

function truncateStepNameForSuffix(name: string, suffix: string): string {
  return [...name].slice(0, MAX_STEP_NAME_LENGTH - [...suffix].length).join('')
}

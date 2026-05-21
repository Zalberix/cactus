import { useControlSteps } from './useControlSteps'

export interface ControlOutcome {
  id: string
  label: string
  color: string
  value?: string
}

export interface SwitchCaseOutcome {
  id: string
  label: string
  value: string
}

export function slugSwitchCase(value: string): string {
  const slug = value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return slug || 'case'
}

export function createSwitchCaseId(labelOrValue: string): string {
  const shortId = Math.random().toString(36).slice(2, 8)
  return `case-${slugSwitchCase(labelOrValue || 'output')}-${shortId}`
}

export function normalizeSwitchCases(cases: unknown): SwitchCaseOutcome[] {
  if (!Array.isArray(cases)) return []

  const result: SwitchCaseOutcome[] = []
  for (const item of cases) {
    if (typeof item === 'string') {
      const value = item.trim()
      if (!value) continue
      result.push({
        id: `case-${slugSwitchCase(value)}`,
        label: value,
        value,
      })
      continue
    }

    if (!item || typeof item !== 'object') continue
    const raw = item as Record<string, unknown>
    const label = typeof raw.label === 'string' ? raw.label.trim() : ''
    const value = typeof raw.value === 'string' ? raw.value.trim() : ''
    const id = typeof raw.id === 'string' ? raw.id.trim() : ''
    if (!id && !label && !value) continue
    result.push({
      id: id || `case-${slugSwitchCase(label || value)}`,
      label: label || value || id,
      value: value || label || id,
    })
  }
  return result
}

export function controlOutcomesForStep(data: {
  controlKind?: string
  controlSettings?: Record<string, unknown>
}): ControlOutcome[] {
  if (data.controlKind === 'switch') {
    const cases = normalizeSwitchCases(data.controlSettings?.cases).map(item => ({
      id: item.id,
      label: item.label,
      color: '#f59e0b',
      value: item.value,
    }))
    return [
      ...cases,
      { id: 'default', label: 'default', color: '#6b7280' },
    ]
  }

  const definition = useControlSteps().find(control => control.kind === data.controlKind)
  return definition?.handles ?? [{ id: 'success', label: '', color: '#22c55e' }]
}

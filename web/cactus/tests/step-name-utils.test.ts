import { describe, expect, it } from 'vitest'
import {
  isStepNameDuplicate,
  nextUniqueStepName,
  normalizeStepName,
} from '../app/components/dag/step-name-utils'

describe('DAG step name utilities', () => {
  it('trims step names before validation and persistence', () => {
    expect(normalizeStepName('  Название  ')).toBe('Название')
  })

  it('creates the next numbered name for an existing base name', () => {
    expect(nextUniqueStepName('Название', ['Название', 'Название 2'])).toBe('Название 3')
  })

  it('uses the next highest suffix when numbered names have gaps', () => {
    expect(nextUniqueStepName('Название', ['Название', 'Название 3'])).toBe('Название 4')
  })

  it('detects duplicate names while ignoring the renamed node', () => {
    expect(isStepNameDuplicate('Название', [
      { id: '1', name: 'Название' },
      { id: '2', name: 'Другой шаг' },
    ], '1')).toBe(false)
    expect(isStepNameDuplicate('Название', [
      { id: '1', name: 'Название' },
      { id: '2', name: 'Другой шаг' },
    ], '2')).toBe(true)
  })
})

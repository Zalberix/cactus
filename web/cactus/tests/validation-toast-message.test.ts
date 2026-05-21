import { describe, expect, it } from 'vitest'
import { validationToastDescription } from '../app/components/dag/validation-toast'

describe('validation toast description', () => {
  it('uses the first validation error message when available', () => {
    expect(validationToastDescription(
      [
        { type: 'missing_mapping', step_id: 12, message: 'Email is required' },
        { type: 'cycle', step_id: 20, message: 'Cycle detected' },
      ],
      'Fallback message',
    )).toBe('Email is required')
  })

  it('falls back to the generic validation message when there are no messages', () => {
    expect(validationToastDescription([], 'Fallback message')).toBe('Fallback message')
    expect(validationToastDescription([{ message: '' }], 'Fallback message')).toBe('Fallback message')
  })
})

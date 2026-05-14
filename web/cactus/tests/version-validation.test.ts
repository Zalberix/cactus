import { describe, expect, it } from 'vitest'
import { normalizeValidationResult } from '../app/composables/useVersions'

describe('version validation', () => {
  it('preserves structured validation errors from the API', () => {
    expect(normalizeValidationResult({
      is_valid: false,
      errors: [
        { type: 'missing_mapping', step_id: 12, field: 'email', message: 'Email is required' },
        'Legacy validation error',
      ],
    })).toEqual({
      isValid: false,
      errors: [
        { type: 'missing_mapping', step_id: 12, field: 'email', message: 'Email is required' },
        { message: 'Legacy validation error' },
      ],
    })
  })
})

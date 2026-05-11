import { describe, expect, it } from 'vitest'
import { isVersionReadOnly } from '../app/components/dag/editor-utils'

describe('workflow version editor header', () => {
  it('marks active versions with runs as read-only', () => {
    const version = { is_active: true, run_count: 3 }
    expect(isVersionReadOnly(version)).toBe(true)
  })
})

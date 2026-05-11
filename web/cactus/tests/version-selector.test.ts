import { describe, expect, it } from 'vitest'
import { sortVersionsForDisplay } from '../app/components/dag/version-utils'

describe('version selector display', () => {
  it('sorts active versions first and newer versions first', () => {
    const sorted = sortVersionsForDisplay([
      { id: 1, version_number: 1, is_active: false },
      { id: 2, version_number: 2, is_active: true },
      { id: 3, version_number: 3, is_active: true },
    ])

    expect(sorted.map(v => v.id)).toEqual([3, 2, 1])
  })
})

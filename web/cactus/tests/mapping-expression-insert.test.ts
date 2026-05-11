import { describe, expect, it } from 'vitest'
import { mappingRecordToEntries, setMappingExpression } from '../app/components/dag/node-editor/mapping-utils'

describe('mapping expression insertion', () => {
  it('inserts clicked schema path into active mapping field', () => {
    const mapping = setMappingExpression({}, 'email', '$.steps.2.output.email')

    expect(mapping.email).toBe('$.steps.2.output.email')
  })

  it('normalizes mapping records to canonical entries', () => {
    expect(mappingRecordToEntries({ email: '$.steps.2.output.email', empty: '' })).toEqual([
      { target: 'email', source: '$.steps.2.output.email' },
    ])
  })
})

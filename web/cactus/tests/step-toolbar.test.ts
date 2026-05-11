import { describe, expect, it } from 'vitest'
import { filterStepCatalog } from '../app/components/dag/step-toolbar-utils'

describe('step toolbar catalog', () => {
  it('searches tasks and control steps together', () => {
    const result = filterStepCatalog([
      { name: 'Email SMTP', category: 'Channels', stepType: 'task', available: true },
      { name: 'IF Condition', category: 'Logic', stepType: 'control', available: true },
    ], 'if')

    expect(result.map(i => i.name)).toEqual(['IF Condition'])
  })
})

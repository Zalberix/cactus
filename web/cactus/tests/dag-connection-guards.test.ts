import { describe, expect, it } from 'vitest'
import { canConnectSteps } from '../app/composables/dag-connection-guards'

describe('dag connection guards', () => {
  it('rejects connecting into start node', () => {
    expect(canConnectSteps({
      source: '2',
      target: '1',
      sourceHandle: 'success',
      nodes: [
        { id: '1', data: { controlKind: 'start' } },
        { id: '2', data: { stepType: 'task' } },
      ],
    })).toBe(false)
  })

  it('allows task success output into normal input', () => {
    expect(canConnectSteps({
      source: '2',
      target: '3',
      sourceHandle: 'success',
      nodes: [
        { id: '2', data: { stepType: 'task' } },
        { id: '3', data: { stepType: 'task' } },
      ],
    })).toBe(true)
  })
})

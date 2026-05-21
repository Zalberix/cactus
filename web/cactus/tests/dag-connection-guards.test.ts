import { describe, expect, it } from 'vitest'
import { canConnectSteps, getControlOutcomeIds } from '../app/composables/dag-connection-guards'
import { controlOutcomesForStep } from '../app/composables/control-outcomes'

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

  it('uses switch case ids from control settings as valid outcomes', () => {
    const switchData = {
      stepType: 'control',
      controlKind: 'switch',
      controlSettings: {
        expression: '$.message.value.type',
        cases: [
          { id: 'case-vip', label: 'VIP', value: 'vip' },
          { id: 'case-regular', label: 'Regular', value: 'regular' },
        ],
      },
    }

    expect(getControlOutcomeIds(switchData)).toEqual(['case-vip', 'case-regular', 'default'])
    expect(controlOutcomesForStep(switchData).map(outcome => outcome.label)).toEqual(['VIP', 'Regular', 'default'])
    expect(canConnectSteps({
      source: '2',
      target: '3',
      sourceHandle: 'case-vip',
      nodes: [
        { id: '2', data: switchData },
        { id: '3', data: { stepType: 'task' } },
      ],
    })).toBe(true)
  })

  it('normalizes legacy string switch cases into stable case outcomes', () => {
    expect(controlOutcomesForStep({
      stepType: 'control',
      controlKind: 'switch',
      controlSettings: {
        expression: '$.message.value.type',
        cases: ['vip'],
      },
    })).toEqual([
      { id: 'case-vip', label: 'vip', color: '#f59e0b', value: 'vip' },
      { id: 'default', label: 'default', color: '#6b7280' },
    ])
  })
})

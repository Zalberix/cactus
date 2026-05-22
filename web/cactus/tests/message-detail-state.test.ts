import { describe, expect, it } from 'vitest'
import {
  normalizeRuntimeStatus,
  runtimeEdgeState,
  runtimeStepMap,
  toCanvasPosition,
} from '../app/composables/useDagViewer'
import type { StepRunDetail } from '../app/composables/useMessages'

describe('message detail state helpers', () => {
  it('normalizes runtime statuses for node rendering', () => {
    expect(normalizeRuntimeStatus('completed')).toBe('completed')
    expect(normalizeRuntimeStatus('done')).toBe('completed')
    expect(normalizeRuntimeStatus('failed')).toBe('failed')
    expect(normalizeRuntimeStatus('error')).toBe('failed')
    expect(normalizeRuntimeStatus('skipped')).toBe('skipped')
    expect(normalizeRuntimeStatus('unexpected')).toBe('pending')
    expect(normalizeRuntimeStatus(undefined)).toBe('pending')
  })

  it('uses backend canvas positions before default positions', () => {
    expect(toCanvasPosition({ x: 120, y: 80 }, 0)).toEqual({ x: 120, y: 80 })
    expect(toCanvasPosition({}, 1)).toEqual({ x: 290, y: 50 })
    expect(toCanvasPosition(undefined, 2)).toEqual({ x: 530, y: 50 })
  })

  it('keys runtime steps by workflow step id', () => {
    const steps: StepRunDetail[] = [
      { id: 8, step_id: 20, status: 'completed' },
      { id: 9, step_id: 30, status: 'running' },
    ]

    const map = runtimeStepMap(steps)

    expect(map.get(20)?.id).toBe(8)
    expect(map.get(30)?.status).toBe('running')
  })

  it('derives semantic runtime state for dependency edges', () => {
    expect(runtimeEdgeState({
      sourceStatus: 'completed',
      sourceOutcome: 'success',
      edgeOutcome: 'success',
      targetStatus: 'completed',
    })).toBe('completed')

    expect(runtimeEdgeState({
      sourceStatus: 'completed',
      sourceOutcome: 'success',
      edgeOutcome: 'success',
      targetStatus: 'running',
    })).toBe('running')

    expect(runtimeEdgeState({
      sourceStatus: 'completed',
      sourceOutcome: 'true',
      edgeOutcome: 'false',
      targetStatus: 'pending',
    })).toBe('pending')

    expect(runtimeEdgeState({
      sourceStatus: 'completed',
      edgeOutcome: 'success',
      targetStatus: 'failed',
    })).toBe('failed')
  })
})

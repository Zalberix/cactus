import { mount } from '@vue/test-utils'
import { computed, nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DagCanvas from '../app/components/dag/DagCanvas.vue'

vi.mock('@vue-flow/core', () => ({
  VueFlow: {
    template: '<div data-testid="vue-flow"><slot name="node-step" id="1" :data="{}" /></div>',
  },
  useVueFlow: () => ({
    screenToFlowCoordinate: ({ x, y }: { x: number, y: number }) => ({ x, y }),
    fitView: vi.fn(),
  }),
}))

vi.mock('@vue-flow/background', () => ({
  Background: {
    template: '<div data-testid="background" />',
  },
}))

vi.mock('@vue-flow/controls', () => ({
  Controls: {
    template: '<div data-testid="controls" />',
  },
}))

function mountCanvas(mode: 'edit' | 'view', props: Record<string, unknown> = {}) {
  return mount(DagCanvas, {
    props: {
      mode,
      nodes: [{ id: '1', type: 'step', position: { x: 0, y: 0 }, data: {} }],
      edges: [],
      ...props,
    },
    global: {
      stubs: {
        StepNode: {
          props: ['readOnly', 'showRuntimeState'],
          template: '<div data-testid="step-node-stub" :data-read-only="String(readOnly)" :data-show-runtime-state="String(showRuntimeState)" />',
        },
        StepEdge: true,
      },
    },
  })
}

describe('DAG canvas read-only mode', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('nextTick', nextTick)
    vi.stubGlobal('onMounted', (fn: () => void) => fn())
  })

  it('passes read-only state to step nodes from canvas mode', () => {
    expect(mountCanvas('view').get('[data-testid="step-node-stub"]').attributes('data-read-only')).toBe('true')
    expect(mountCanvas('edit').get('[data-testid="step-node-stub"]').attributes('data-read-only')).toBe('false')
  })

  it('keeps runtime state indicators opt-in instead of tying them to read-only mode', () => {
    expect(mountCanvas('view').get('[data-testid="step-node-stub"]').attributes('data-show-runtime-state')).toBe('false')
    expect(mountCanvas('view', { showRuntimeState: true }).get('[data-testid="step-node-stub"]').attributes('data-show-runtime-state')).toBe('true')
  })
})

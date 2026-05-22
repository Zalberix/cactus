import { mount } from '@vue/test-utils'
import { computed, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StepEdge from '../app/components/dag/StepEdge.vue'

vi.mock('@vue-flow/core', () => ({
  BaseEdge: {
    name: 'BaseEdge',
    props: ['id', 'path', 'interactionWidth', 'style'],
    template: '<path data-testid="base-edge" :data-id="id" :data-path="path" />',
  },
  EdgeLabelRenderer: {
    template: '<div data-testid="edge-label-renderer"><slot /></div>',
  },
  getSmoothStepPath: () => ['M 0 0 L 100 0', 50, 20],
}))

function mountEdge(props: Record<string, unknown> = {}) {
  return mount(StepEdge, {
    props: {
      id: 'e-1-2-false',
      sourceX: 0,
      sourceY: 0,
      targetX: 100,
      targetY: 0,
      sourcePosition: 'right',
      targetPosition: 'left',
      source: '1',
      target: '2',
      sourceNode: { id: '1' },
      targetNode: { id: '2' },
      sourceHandleId: 'false',
      label: 'false',
      type: 'step',
      selected: false,
      markerStart: '',
      markerEnd: '',
      data: {},
      events: {},
      ...props,
    },
  })
}

describe('DAG StepEdge UX', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
  })

  it('does not render output names as labels on connections', () => {
    const wrapper = mountEdge()

    expect(wrapper.text()).not.toContain('false')
  })

})

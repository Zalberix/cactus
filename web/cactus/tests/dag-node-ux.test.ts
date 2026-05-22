import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, defineComponent, ref } from 'vue'
import StepNode from '../app/components/dag/StepNode.vue'

const currentNode = vi.hoisted(() => ({
  value: {
    id: '1',
    selected: false,
    data: {},
  } as any,
}))

vi.mock('@vue-flow/core', () => ({
  Handle: {
    props: ['id', 'type', 'position'],
    template: '<span data-testid="handle" :data-handle-id="id" :data-handle-type="type" :data-position="position" />',
  },
  Position: {
    Left: 'left',
    Right: 'right',
  },
  useNode: () => ({ node: currentNode.value }),
}))

vi.mock('../app/composables/useControlSteps', () => ({
  useControlSteps: () => [
    {
      kind: 'condition',
      name: 'IF Condition',
      category: 'Logic',
      icon: 'git-branch',
      color: '#ef4444',
      settingsSchema: {},
      handles: [
        { id: 'true', label: 'true', color: '#22c55e' },
        { id: 'false', label: 'false', color: '#ef4444' },
      ],
    },
  ],
}))

vi.mock('../app/composables/control-outcomes', () => ({
  controlOutcomesForStep: () => [
    { id: 'true', label: 'true', color: '#22c55e' },
    { id: 'false', label: 'false', color: '#ef4444' },
  ],
}))

const DropdownStub = {
  template: '<div><slot /></div>',
}

const DropdownItemStub = defineComponent({
  emits: ['select'],
  setup(_, { emit }) {
    const selectDefaultPrevented = ref<string>()

    function onClick() {
      const event = new Event('select', { cancelable: true })
      emit('select', event)
      selectDefaultPrevented.value = String(event.defaultPrevented)
    }

    return { onClick, selectDefaultPrevented }
  },
  template: '<button type="button" data-testid="rename-node" :data-select-default-prevented="selectDefaultPrevented" @click="onClick"><slot /></button>',
})

function mountNode(props: Record<string, unknown> = {}) {
  return mount(StepNode, {
    props,
    global: {
      stubs: {
        DropdownMenu: DropdownStub,
        DropdownMenuContent: DropdownStub,
        DropdownMenuTrigger: DropdownStub,
        DropdownMenuItem: DropdownItemStub,
      },
    },
  })
}

function setNode(data: Record<string, unknown>, id = '12') {
  currentNode.value = {
    id,
    selected: false,
    data: {
      label: 'SMTP Worker',
      stepType: 'task',
      status: 'pending',
      ...data,
    },
  }
}

describe('DAG StepNode UX', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
    setNode({})
  })

  it('renders the step name outside the node block and omits runtime status', () => {
    const wrapper = mountNode()

    expect(wrapper.get('[data-testid="step-node-block"]').text()).not.toContain('SMTP Worker')
    expect(wrapper.get('[data-testid="step-node-label"]').text()).toContain('SMTP Worker')
    expect(wrapper.find('[data-testid="step-node-status"]').exists()).toBe(false)
  })

  it('shows hover actions for regular nodes and emits delete with the node id', async () => {
    const wrapper = mountNode()
    const actions = wrapper.get('[data-testid="node-hover-actions"]')
    const actionClasses = actions.attributes('class')?.split(/\s+/) ?? []

    expect(wrapper.get('[data-testid="delete-node"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="more-node"]').exists()).toBe(true)
    expect(actionClasses).toContain('pointer-events-auto')
    expect(actionClasses).toContain('hover:opacity-100')
    expect(actionClasses).not.toContain('pointer-events-none')
    expect(wrapper.get('[data-testid="rename-node"]').text()).toContain('editor.renameStep')

    await wrapper.get('[data-testid="delete-node"]').trigger('click')

    expect(wrapper.emitted('delete')).toEqual([['12']])
  })

  it('hides edit actions when rendered read-only', () => {
    const wrapper = mountNode({ readOnly: true })

    expect(wrapper.find('[data-testid="node-hover-actions"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="delete-node"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="more-node"]').exists()).toBe(false)
  })

  it('emits rename from the node action menu', async () => {
    const wrapper = mountNode()

    await wrapper.get('[data-testid="rename-node"]').trigger('click')

    expect(wrapper.get('[data-testid="rename-node"]').attributes('data-select-default-prevented')).toBe('false')
    expect(wrapper.emitted('rename')).toEqual([['12']])
  })

  it('uses a neutral node border instead of per-step accent borders', () => {
    setNode({
      workTypeMeta: { icon: 'mail', color: '#ef4444' },
    })

    const wrapper = mountNode()

    expect(wrapper.get('[data-testid="step-node-block"]').attributes('class')).toContain('border-slate-300')
    expect(wrapper.get('[data-testid="step-node-block"]').attributes('style') ?? '').not.toContain('border-color')
  })

  it('places output names above the outgoing handles', () => {
    setNode({
      stepType: 'control',
      controlKind: 'condition',
    })

    const wrapper = mountNode()
    const labels = wrapper.findAll('[data-testid="output-handle-label"]')

    expect(labels.map(label => label.text())).toEqual(['true', 'false'])
    for (const label of labels) {
      expect(label.attributes('style')).toContain('translateY(calc(-100% - 6px))')
    }
  })

  it('does not show actions for the protected start node', () => {
    setNode({
      label: 'System Trigger',
      stepType: 'control',
      controlKind: 'start',
    }, '1')

    const wrapper = mountNode()

    expect(wrapper.find('[data-testid="delete-node"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="more-node"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="step-node-label"]').text()).toContain('System Trigger')
  })

  it('shows red validation state and lower-right validation details', () => {
    setNode({
      validationErrors: [
        { type: 'missing_mapping', step_id: 12, message: 'Email is required' },
        { type: 'missing_dependency', step_id: 12, message: 'Connect a dependency' },
      ],
    })

    const wrapper = mountNode()

    expect(wrapper.get('[data-testid="step-node-block"]').classes().join(' ')).toContain('border-red')
    expect(wrapper.get('[data-testid="validation-indicator"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="validation-tooltip"]').text()).toContain('Email is required')
    expect(wrapper.get('[data-testid="validation-tooltip"]').text()).toContain('Connect a dependency')
  })
})

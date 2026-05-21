import { mount } from '@vue/test-utils'
import { computed, defineComponent, h, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ControlParametersPanel from '../app/components/dag/node-editor/ControlParametersPanel.vue'
import SwitchControlForm from '../app/components/dag/node-editor/SwitchControlForm.vue'

const ButtonStub = defineComponent({
  name: 'Button',
  setup(_, { slots, attrs }) {
    return () => h('button', attrs, slots.default?.())
  },
})

const InputStub = defineComponent({
  name: 'Input',
  props: ['modelValue'],
  emits: ['update:modelValue', 'focus'],
  setup(props, { emit, attrs }) {
    return () => h('input', {
      ...attrs,
      value: props.modelValue,
      onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
      onFocus: () => emit('focus'),
    })
  },
})

const SelectStub = defineComponent({
  name: 'Select',
  props: ['modelValue'],
  emits: ['update:modelValue'],
  setup(props, { emit, slots }) {
    return () => h('select', {
      value: props.modelValue,
      onChange: (event: Event) => emit('update:modelValue', (event.target as HTMLSelectElement).value),
    }, slots.default?.())
  },
})

function baseControlStep(overrides: Record<string, unknown> = {}) {
  return {
    label: 'Condition',
    stepType: 'control',
    controlKind: 'condition',
    controlSettings: {},
    status: 'pending',
    ...overrides,
  }
}

describe('control parameters panel', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  })

  it('inserts expressions into the active condition side', async () => {
    const wrapper = mount(ControlParametersPanel, {
      props: {
        stepData: baseControlStep({
          controlSettings: { left: '', operator: 'eq', right: '' },
        }),
      },
      global: { stubs: panelStubs() },
    })

    await wrapper.get('[data-testid="condition-left"]').trigger('focus')
    wrapper.vm.insertExpression('$.message.value.age')
    await wrapper.get('[data-testid="condition-right"]').trigger('focus')
    wrapper.vm.insertExpression('$.steps.2.output.limit')
    await wrapper.get('[data-testid="save-control-settings"]').trigger('click')

    expect(wrapper.emitted('saveControlSettings')).toEqual([[
      { left: '$.message.value.age', operator: 'eq', right: '$.steps.2.output.limit' },
    ]])
  })
})

describe('switch control form', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  })

  it('adds a case, keeps its id stable across label and value edits, and emits settings', async () => {
    const wrapper = mount(SwitchControlForm, {
      props: {
        modelValue: {
          expression: '$.message.value.type',
          cases: [],
        },
      },
      global: { stubs: panelStubs() },
    })

    await wrapper.get('[data-testid="add-switch-case"]').trigger('click')
    const firstEmission = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as { cases: Array<{ id: string }> }
    const caseID = firstEmission.cases[0].id

    await wrapper.get(`[data-testid="switch-case-label-${caseID}"]`).setValue('VIP')
    await wrapper.get(`[data-testid="switch-case-value-${caseID}"]`).setValue('vip')

    const lastEmission = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as { cases: Array<{ id: string, label: string, value: string }> }
    expect(lastEmission.cases[0]).toEqual({ id: caseID, label: 'VIP', value: 'vip' })
  })
})

function panelStubs() {
  return {
    Button: ButtonStub,
    Input: InputStub,
    ScrollArea: { template: '<div><slot /></div>' },
    DynamicSettingsForm: { template: '<div />' },
    ExpressionField: InputStub,
    Label: { template: '<label><slot /></label>' },
    Select: SelectStub,
    SelectContent: { template: '<div><slot /></div>' },
    SelectItem: { props: ['value'], template: '<option :value="value"><slot /></option>' },
    SelectTrigger: { template: '<button v-bind="$attrs"><slot /></button>' },
    SelectValue: { template: '<span v-bind="$attrs"><slot /></span>' },
  }
}

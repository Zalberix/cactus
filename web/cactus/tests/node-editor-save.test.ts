import { mount } from '@vue/test-utils'
import { computed, defineComponent, h, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ParametersPanel from '../app/components/dag/node-editor/ParametersPanel.vue'
import NodeEditor from '../app/components/dag/node-editor/NodeEditor.vue'

const ButtonStub = defineComponent({
  name: 'Button',
  setup(_, { slots, attrs }) {
    return () => h('button', attrs, slots.default?.())
  },
})

function baseStepData(overrides: Record<string, unknown> = {}) {
  return {
    label: 'SMTP',
    stepType: 'task',
    status: 'pending',
    config: {},
    inputMapping: {},
    settingsSchema: { type: 'object', properties: {} },
    inputSchema: { type: 'object', properties: {} },
    ...overrides,
  }
}

describe('node editor save actions', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  })

  it('clicking settings save emits only saveSettings', async () => {
    const wrapper = mount(ParametersPanel, {
      props: {
        stepData: baseStepData({
          config: { host: 'smtp.local' },
          settingsSchema: {
            type: 'object',
            properties: { host: { type: 'string', required: true } },
          },
        }),
      },
      global: {
        stubs: panelStubs(),
      },
    })

    await wrapper.get('[data-testid="save-settings"]').trigger('click')

    expect(wrapper.emitted('saveSettings')).toEqual([[{ host: 'smtp.local' }]])
    expect(wrapper.emitted('saveInputMapping')).toBeUndefined()
  })

  it('clicking input mapping save emits only saveInputMapping', async () => {
    const wrapper = mount(ParametersPanel, {
      props: {
        stepData: baseStepData({
          inputMapping: { to: '$.message.value.email' },
          inputSchema: {
            type: 'object',
            properties: { to: { type: 'string', required: true } },
          },
        }),
      },
      global: {
        stubs: panelStubs(),
      },
    })

    await wrapper.get('[data-testid="save-input-mapping"]').trigger('click')

    expect(wrapper.emitted('saveInputMapping')).toEqual([[[{ target: 'to', source: '$.message.value.email' }]]])
    expect(wrapper.emitted('saveSettings')).toBeUndefined()
  })

  it('invalid required settings show feedback and emit no save', async () => {
    const wrapper = mount(ParametersPanel, {
      props: {
        stepData: baseStepData({
          settingsSchema: {
            type: 'object',
            properties: { host: { type: 'string', required: true } },
          },
        }),
      },
      global: {
        stubs: panelStubs(),
      },
    })

    await wrapper.get('[data-testid="save-settings"]').trigger('click')

    expect(wrapper.emitted('saveSettings')).toBeUndefined()
    expect(wrapper.get('[data-testid="settings-validation-errors"]').text()).toContain('host')
  })

  it('NodeEditor forwards split save events with node id', async () => {
    const wrapper = mount(NodeEditor, {
      props: nodeEditorProps(),
      global: {
        stubs: nodeEditorStubs({
          ParametersPanel: {
            emits: ['saveSettings', 'saveInputMapping', 'createWorkflowInput'],
            template: `
              <div>
                <button data-testid="emit-settings" @click="$emit('saveSettings', { host: 'smtp.local' })">settings</button>
                <button data-testid="emit-mapping" @click="$emit('saveInputMapping', [{ target: 'to', source: '$.message.value.email' }])">mapping</button>
              </div>
            `,
          },
        }),
      },
    })

    await wrapper.get('[data-testid="emit-settings"]').trigger('click')
    await wrapper.get('[data-testid="emit-mapping"]').trigger('click')

    expect(wrapper.emitted('saveSettings')).toEqual([['12', { host: 'smtp.local' }]])
    expect(wrapper.emitted('saveInputMapping')).toEqual([['12', [{ target: 'to', source: '$.message.value.email' }]]])
  })

  it('NodeEditor forwards control settings saves with node id', async () => {
    const wrapper = mount(NodeEditor, {
      props: nodeEditorProps({
        stepType: 'control',
        controlKind: 'condition',
        controlSettings: { left: '$.message.value.age', operator: 'gte', right: '18' },
      }),
      global: {
        stubs: nodeEditorStubs({
          ControlParametersPanel: {
            emits: ['saveControlSettings'],
            template: `
              <div>
                <button data-testid="emit-control-settings" @click="$emit('saveControlSettings', { left: '$.message.value.age', operator: 'gte', right: '18' })">control</button>
              </div>
            `,
          },
        }),
      },
    })

    await wrapper.get('[data-testid="emit-control-settings"]').trigger('click')

    expect(wrapper.emitted('saveControlSettings')).toEqual([['12', { left: '$.message.value.age', operator: 'gte', right: '18' }]])
    expect(wrapper.findComponent({ name: 'ParametersPanel' }).exists()).toBe(false)
  })

  it('NodeEditor renders only the built-in close button', () => {
    const wrapper = mount(NodeEditor, {
      props: nodeEditorProps(),
      global: {
        stubs: nodeEditorStubs(),
      },
    })

    expect(wrapper.findAll('button').length).toBeLessThanOrEqual(1)
  })
})

function panelStubs() {
  return {
    Button: ButtonStub,
    ScrollArea: { template: '<div><slot /></div>' },
    Tabs: { template: '<div><slot /></div>' },
    TabsContent: { template: '<div><slot /></div>' },
    TabsList: { template: '<div><slot /></div>' },
    TabsTrigger: { template: '<button><slot /></button>' },
    DynamicSettingsForm: { template: '<div />' },
    ExpressionField: { template: '<input />' },
    Label: { template: '<label><slot /></label>' },
  }
}

function nodeEditorProps(stepOverrides: Record<string, unknown> = {}) {
  return {
    open: true,
    nodeId: '12',
    workflowId: 42,
    versionId: 99,
    allEdges: [],
    allNodes: [{
      id: '12',
      data: baseStepData({
        workTypeName: 'SMTP',
        workTypeCode: 'smtp',
        ...stepOverrides,
      }),
    }],
  }
}

function nodeEditorStubs(extra: Record<string, unknown> = {}) {
  return {
    Badge: { template: '<span><slot /></span>' },
    Dialog: { template: '<div><slot /></div>' },
    DialogContent: { template: '<section><slot /><button aria-label="Close">Close</button></section>' },
    InputPanel: { template: '<div />' },
    ParametersPanel: { template: '<div />' },
    ControlParametersPanel: { template: '<div />' },
    ...extra,
  }
}

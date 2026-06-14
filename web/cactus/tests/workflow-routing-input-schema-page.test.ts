import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, h, reactive, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import NewInputSchemaPage from '../app/pages/org/[orgId]/workflows/[workflowId]/routing/input-schemas/new.vue'
import EditInputSchemaPage from '../app/pages/org/[orgId]/workflows/[workflowId]/routing/input-schemas/[inputSchemaId]/edit.vue'
import { getErrorMessage } from '../app/utils/errors'

const toastMock = vi.hoisted(() => vi.fn())
const routerPushMock = vi.hoisted(() => vi.fn())
const routingMock = vi.hoisted(() => ({
  fetchInputSchema: vi.fn(),
  createInputSchema: vi.fn(),
  updateInputSchema: vi.fn(),
  updateInputSchemaStatus: vi.fn(),
  setInputSchemaDefault: vi.fn(),
}))

vi.mock('~/components/ui/toast/use-toast', () => ({
  toast: toastMock,
}))

describe('workflow routing input schema pages', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    toastMock.mockClear()
    routerPushMock.mockClear()
    for (const mock of Object.values(routingMock)) mock.mockReset()

    routingMock.createInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object', properties: {} },
      status: 'active',
      is_default: false,
    })
    routingMock.updateInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object', properties: {} },
      status: 'active',
      is_default: true,
    })
    routingMock.updateInputSchemaStatus.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object', properties: {} },
      status: 'active',
      is_default: true,
    })
    routingMock.setInputSchemaDefault.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object', properties: {} },
      status: 'active',
      is_default: true,
    })

    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('reactive', reactive)
    vi.stubGlobal('onMounted', (callback: () => void) => callback())
    vi.stubGlobal('getErrorMessage', getErrorMessage)
    vi.stubGlobal('useRouter', () => ({
      push: routerPushMock,
    }))
    vi.stubGlobal('useI18n', () => ({
      t: translate,
    }))
    vi.stubGlobal('useWorkflowRouting', () => routingMock)
  })

  it('creates a versioned input schema from table fields', async () => {
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42' },
    }))

    const wrapper = mount(NewInputSchemaPage, {
      global: { stubs: pageStubs() },
    })
    await flushPromises()

    await wrapper.get('[data-testid="routing-schema-code-input"]').setValue('public')
    await wrapper.get('[data-testid="routing-schema-add-field"]').trigger('click')
    await wrapper.get('[data-testid="routing-schema-field-name-input"]').setValue('email')
    await wrapper.get('[data-testid="routing-schema-field-required-checkbox"]').trigger('click')
    await wrapper.get('[data-testid="routing-schema-field-description-input"]').setValue('Recipient address')
    await wrapper.get('[data-testid="routing-schema-save-field"]').trigger('click')

    expect(wrapper.find('[data-testid="workflow-input-row-email"]').exists()).toBe(true)

    await wrapper.get('[data-testid="routing-schema-save"]').trigger('click')
    await flushPromises()

    expect(routingMock.createInputSchema).toHaveBeenCalledWith(42, {
      code: 'public',
      status: 'active',
      is_default: false,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true, description: 'Recipient address' },
        },
      },
    })
    expect(routerPushMock).toHaveBeenCalledWith('/org/7/workflows/42/routing')
  })

  it('renders field dialog controls with the same row alignment as the page filter', async () => {
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42' },
    }))

    const wrapper = mount(NewInputSchemaPage, {
      global: { stubs: pageStubs() },
    })
    await flushPromises()

    await wrapper.get('[data-testid="routing-schema-add-field"]').trigger('click')

    const fieldForm = wrapper.get('[data-testid="routing-schema-field-form"]')
    const fieldGrid = wrapper.get('[data-testid="routing-schema-field-controls"]')
    const fieldLabels = wrapper.findAll('[data-testid="routing-schema-field-controls"] > label')
    expect(fieldLabels).toHaveLength(3)
    expect(fieldForm.element.tagName).toBe('DIV')
    expect(fieldForm.find('select').exists()).toBe(false)
    expect(fieldGrid.classes()).toContain('md:items-end')
    expect(fieldGrid.classes()).not.toContain('md:items-start')
    expect(fieldLabels[0].classes()).toContain('block')
    expect(fieldLabels[1].classes()).toContain('block')
    expect(fieldLabels[0].classes()).toContain('space-y-1.5')
    expect(fieldLabels[1].classes()).toContain('space-y-1.5')
    expect(wrapper.get('[data-testid="routing-schema-field-type-select"]').classes()).toContain('h-10')
    expect(fieldLabels[2].classes()).toContain('h-10')
  })

  it('edits an existing versioned input schema from table fields', async () => {
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42', inputSchemaId: '11' },
    }))
    routingMock.fetchInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
        },
      },
      status: 'active',
      is_default: true,
    })

    const wrapper = mount(EditInputSchemaPage, {
      global: { stubs: pageStubs() },
    })
    await flushPromises()

    expect(routingMock.fetchInputSchema).toHaveBeenCalledWith(11)
    expect(wrapper.find('[data-testid="workflow-input-row-email"]').exists()).toBe(true)

    await wrapper.get('[data-testid="routing-schema-save"]').trigger('click')
    await flushPromises()

    expect(routingMock.updateInputSchema).toHaveBeenCalledWith(11, {
      code: 'public',
      is_default: true,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
        },
      },
    })
    expect(routingMock.updateInputSchemaStatus).toHaveBeenCalledWith(11, 'active')
    expect(routingMock.setInputSchemaDefault).toHaveBeenCalledWith(11)
    expect(routerPushMock).toHaveBeenCalledWith('/org/7/workflows/42/routing')
  })

  it('renames an existing input schema field from the edit form', async () => {
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42', inputSchemaId: '11' },
    }))
    routingMock.fetchInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
        },
      },
      status: 'active',
      is_default: false,
    })

    const wrapper = mount(EditInputSchemaPage, {
      global: { stubs: pageStubs() },
    })
    await flushPromises()

    await wrapper.get('[data-testid="routing-schema-edit-field-email"]').trigger('click')
    const nameInput = wrapper.get('[data-testid="routing-schema-field-name-input"]')

    expect((nameInput.element as HTMLInputElement).readOnly).toBe(false)

    await nameInput.setValue('contactEmail')
    await wrapper.get('[data-testid="routing-schema-field-description-input"]').setValue('Contact address')
    await wrapper.get('[data-testid="routing-schema-save-field"]').trigger('click')

    expect(wrapper.find('[data-testid="workflow-input-row-email"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="workflow-input-row-contactEmail"]').exists()).toBe(true)

    await wrapper.get('[data-testid="routing-schema-save"]').trigger('click')
    await flushPromises()

    expect(routingMock.updateInputSchema).toHaveBeenCalledWith(11, {
      code: 'public',
      is_default: false,
      schema_json: {
        type: 'object',
        properties: {
          contactEmail: { type: 'string', required: true, description: 'Contact address' },
        },
      },
    })
  })
})

function pageStubs() {
  const passthrough = defineComponent({
    setup(_, { slots }) {
      return () => h('div', slots.default?.())
    },
  })

  return {
    Badge: { template: '<span><slot /></span>' },
    Button: defineComponent({
      props: {
        disabled: { type: Boolean, default: false },
      },
      setup(props, { attrs, slots }) {
        return () => h('button', {
          ...attrs,
          disabled: props.disabled,
          onClick: (event: MouseEvent) => {
            const clickHandler = attrs.onClick
            if (typeof clickHandler === 'function') clickHandler(event)
            const type = attrs.type === undefined ? 'submit' : String(attrs.type)
            if (!event.defaultPrevented && type === 'submit') {
              ;(event.currentTarget as HTMLButtonElement).form?.dispatchEvent(
                new Event('submit', { bubbles: true, cancelable: true }),
              )
            }
          },
        }, slots.default?.())
      },
    }),
    Checkbox: defineComponent({
      props: {
        modelValue: { type: Boolean, default: false },
        disabled: { type: Boolean, default: false },
      },
      emits: ['update:modelValue'],
      setup(props, { attrs, emit }) {
        return () => h('button', {
          ...attrs,
          disabled: props.disabled,
          type: 'button',
          role: 'checkbox',
          'aria-checked': props.modelValue,
          onClick: () => {
            if (!props.disabled) emit('update:modelValue', !props.modelValue)
          },
        })
      },
    }),
    Card: passthrough,
    CardContent: passthrough,
    CardDescription: { template: '<p><slot /></p>' },
    CardHeader: passthrough,
    CardTitle: { template: '<h2><slot /></h2>' },
    Dialog: passthrough,
    DialogContent: passthrough,
    DialogDescription: { template: '<p><slot /></p>' },
    DialogFooter: passthrough,
    DialogHeader: passthrough,
    DialogTitle: { template: '<h2><slot /></h2>' },
    Input: defineComponent({
      props: {
        modelValue: { type: [String, Number], default: '' },
        disabled: { type: Boolean, default: false },
        readonly: { type: Boolean, default: false },
      },
      emits: ['update:modelValue'],
      setup(props, { attrs, emit }) {
        return () => h('input', {
          ...attrs,
          disabled: props.disabled,
          readonly: props.readonly,
          value: props.modelValue,
          onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
        })
      },
    }),
    Table: passthrough,
    TableBody: passthrough,
    TableCell: passthrough,
    TableEmpty: passthrough,
    TableHead: passthrough,
    TableHeader: passthrough,
    TableRow: passthrough,
  }
}

function translate(key: string, params?: Record<string, unknown>) {
  const translations: Record<string, string> = {
    'common.back': 'Back',
    'common.cancel': 'Cancel',
    'common.loading': 'Loading',
    'common.save': 'Save',
    'workflowInputs.actions': 'Actions',
    'workflowInputs.addInput': 'Add input',
    'workflowInputs.deleteConfirmBody': 'Mappings that use this input will become invalid until updated.',
    'workflowInputs.deleteConfirmTitle': `Delete ${params?.name}?`,
    'workflowInputs.description': 'Description',
    'workflowInputs.editInput': 'Edit input',
    'workflowInputs.empty': 'No workflow inputs declared.',
    'workflowInputs.name': 'Name',
    'workflowInputs.noDescription': 'No description',
    'workflowInputs.optional': 'Optional',
    'workflowInputs.required': 'Required',
    'workflowInputs.type': 'Type',
    'workflowRouting.createInputSchema': 'Create input schema',
    'workflowRouting.createdInputSchema': 'Input schema created',
    'workflowRouting.editInputSchema': 'Edit input schema',
    'workflowRouting.errorCreateInputSchema': 'Failed to create input schema',
    'workflowRouting.fieldDefaultSchema': 'Default schema',
    'workflowRouting.fieldEnabled': 'Enabled',
    'workflowRouting.fieldSchemaCode': 'Schema code',
    'workflowRouting.fieldStatus': 'Status',
    'workflowRouting.inputSchemasDescription': 'Versioned public request contracts used by message routing.',
    'workflowRouting.placeholderSchemaCode': 'schema code',
    'workflowRouting.statusActive': 'Active',
    'workflowRouting.statusDraft': 'Draft',
    'workflowRouting.updateInputSchema': 'Update schema',
    'workflowRouting.updatedInputSchema': 'Input schema updated',
  }
  return translations[key] ?? key
}

import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, h, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import WorkflowVersionTable from '../app/components/dag/WorkflowVersionTable.vue'
import WorkflowOverviewPage from '../app/pages/org/[orgId]/workflows/[workflowId]/index.vue'
import { getErrorMessage } from '../app/utils/errors'

const toastMock = vi.hoisted(() => vi.fn())
const workflowMock = vi.hoisted(() => ({
  fetchWorkflow: vi.fn(),
  updateWorkflow: vi.fn(),
  fetchWorkflowInputSchema: vi.fn(),
  upsertWorkflowInputSchemaField: vi.fn(),
  deleteWorkflowInputSchemaField: vi.fn(),
}))
const versionsMock = vi.hoisted(() => ({
  fetchVersionSummaries: vi.fn(),
  createVersion: vi.fn(),
  activateVersion: vi.fn(),
  deactivateVersion: vi.fn(),
  deleteVersion: vi.fn(),
}))
const routingMock = vi.hoisted(() => ({
  fetchInputSchemas: vi.fn(),
  fetchCompatibilities: vi.fn(),
}))

vi.mock('~/components/ui/toast/use-toast', () => ({
  toast: toastMock,
}))

describe('workflow version supported schemas', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    toastMock.mockClear()
    for (const mock of Object.values(workflowMock)) mock.mockReset()
    for (const mock of Object.values(versionsMock)) mock.mockReset()
    for (const mock of Object.values(routingMock)) mock.mockReset()

    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('onMounted', (callback: () => void) => callback())
    vi.stubGlobal('getErrorMessage', getErrorMessage)
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42' },
    }))
    vi.stubGlobal('useRouter', () => ({
      push: vi.fn(),
    }))
    vi.stubGlobal('useI18n', () => ({
      t: translate,
    }))
    vi.stubGlobal('useWorkflows', () => workflowMock)
    vi.stubGlobal('useVersions', () => versionsMock)
    vi.stubGlobal('useWorkflowRouting', () => routingMock)
  })

  it('renders supported input schema versions in the version table', () => {
    const wrapper = mount(WorkflowVersionTable, {
      props: {
        versions: [{
          id: 101,
          workflow_id: 42,
          name: 'Current',
          version_number: 3,
          is_valid: true,
          is_active: true,
          traffic_weight: 100,
          is_control_group: false,
          run_count: 0,
          created_at: '',
        }],
        orgId: 7,
        workflowId: 42,
        supportedSchemaLabelsByVersionId: {
          101: ['partner v2', 'public v1'],
        },
      },
      global: {
        stubs: tableStubs(),
      },
    })

    expect(wrapper.text()).toContain('Supported schema versions')
    expect(wrapper.text()).toContain('partner v2')
    expect(wrapper.text()).toContain('public v1')
  })

  it('loads routing compatibilities and passes active schema labels to the overview table', async () => {
    workflowMock.fetchWorkflow.mockResolvedValue({
      id: 42,
      name: 'Notify customer',
      system_id: 3,
      priority: 1,
    })
    versionsMock.fetchVersionSummaries.mockResolvedValue([{
      id: 101,
      workflow_id: 42,
      name: 'Current',
      version_number: 3,
      is_valid: true,
      is_active: true,
      traffic_weight: 100,
      is_control_group: false,
      run_count: 0,
      created_at: '',
    }])
    routingMock.fetchInputSchemas.mockResolvedValue([{
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object' },
      status: 'active',
      is_default: true,
    }, {
      id: 12,
      workflow_id: 42,
      code: 'partner',
      version_number: 2,
      schema_json: { type: 'object' },
      status: 'active',
      is_default: false,
    }, {
      id: 13,
      workflow_id: 42,
      code: 'internal',
      version_number: 9,
      schema_json: { type: 'object' },
      status: 'active',
      is_default: false,
    }])
    routingMock.fetchCompatibilities.mockResolvedValue([{
      id: 31,
      workflow_version_id: 101,
      workflow_input_schema_id: 11,
      compatibility_type: 'native',
      default_values: {},
      is_active: true,
      is_default_route: true,
    }, {
      id: 32,
      workflow_version_id: 101,
      workflow_input_schema_id: 12,
      compatibility_type: 'adapter',
      default_values: {},
      is_active: true,
      is_default_route: false,
    }, {
      id: 33,
      workflow_version_id: 101,
      workflow_input_schema_id: 13,
      compatibility_type: 'native',
      default_values: {},
      is_active: false,
      is_default_route: false,
    }])

    const wrapper = mount(WorkflowOverviewPage, {
      global: {
        stubs: {
          ...tableStubs(),
          WorkflowVersionCreateMenu: true,
          WorkflowTokensDialog: true,
          WorkflowSchemaDialog: true,
          EmptyState: true,
          Input: inputStub(),
        },
      },
    })
    await flushPromises()

    expect(routingMock.fetchInputSchemas).toHaveBeenCalledWith(42)
    expect(routingMock.fetchCompatibilities).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('public v1')
    expect(wrapper.text()).toContain('partner v2')
    expect(wrapper.text()).not.toContain('internal v9')
  })
})

function tableStubs() {
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
        return () => h('button', { ...attrs, disabled: props.disabled }, slots.default?.())
      },
    }),
    DropdownMenu: passthrough,
    DropdownMenuContent: passthrough,
    DropdownMenuItem: passthrough,
    DropdownMenuTrigger: passthrough,
    NuxtLink: defineComponent({
      props: {
        to: { type: String, required: true },
      },
      setup(props, { slots }) {
        return () => h('a', { href: props.to }, slots.default?.())
      },
    }),
  }
}

function inputStub() {
  return defineComponent({
    props: {
      modelValue: { type: [String, Number], default: '' },
    },
    emits: ['update:modelValue'],
    setup(props, { attrs, emit }) {
      return () => h('input', {
        ...attrs,
        value: props.modelValue,
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
      })
    },
  })
}

function translate(key: string) {
  const translations: Record<string, string> = {
    'common.loading': 'Loading',
    'editor.valid': 'Valid',
    'status.active': 'Active',
    'status.inactive': 'Inactive',
    'workflowRouting.nav': 'Routing',
    'workflowVersions.actions.activate': 'Activate',
    'workflowVersions.actions.deactivate': 'Deactivate',
    'workflowVersions.actions.delete': 'Delete',
    'workflowVersions.actions.workflowInputSchema': 'Workflow input schema',
    'workflowVersions.createMenu.create': 'Create Version',
    'workflowVersions.editing.editable': 'Editable',
    'workflowVersions.editing.readOnly': 'Read-only',
    'workflowVersions.headers.editing': 'Editing',
    'workflowVersions.headers.name': 'Name',
    'workflowVersions.headers.runs': 'Runs',
    'workflowVersions.headers.status': 'Status',
    'workflowVersions.headers.supportedSchemas': 'Supported schema versions',
    'workflowVersions.headers.traffic': 'Traffic',
    'workflowVersions.status.control': 'Control',
    'workflowVersions.status.validationRequired': 'Validation required',
    'workflowVersions.tokens': 'Tokens',
    'workflowVersions.traffic': 'Traffic',
    'workflowVersions.versionFallback': 'Version',
  }
  return translations[key] ?? key
}

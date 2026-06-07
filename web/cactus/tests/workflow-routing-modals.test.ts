import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, h, reactive, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RoutingCompatibilityDialog from '../app/components/dag/routing/RoutingCompatibilityDialog.vue'
import RoutingExperimentDialog from '../app/components/dag/routing/RoutingExperimentDialog.vue'
import RoutingInputMapperDialog from '../app/components/dag/routing/RoutingInputMapperDialog.vue'
import RoutingScopeDialog from '../app/components/dag/routing/RoutingScopeDialog.vue'
import RoutingVariantDialog from '../app/components/dag/routing/RoutingVariantDialog.vue'
import WorkflowRoutingPage from '../app/pages/org/[orgId]/workflows/[workflowId]/routing/index.vue'
import { getErrorMessage } from '../app/utils/errors'

const toastMock = vi.hoisted(() => vi.fn())
const routerPushMock = vi.hoisted(() => vi.fn())
const routeLeaveMock = vi.hoisted(() => vi.fn())
const routingMock = vi.hoisted(() => ({
  fetchInputSchemas: vi.fn(),
  fetchInputMappers: vi.fn(),
  fetchCompatibilities: vi.fn(),
  fetchExperiments: vi.fn(),
  fetchExperimentScopes: vi.fn(),
  fetchExperimentVariants: vi.fn(),
  createInputSchema: vi.fn(),
  updateInputSchema: vi.fn(),
  updateInputSchemaStatus: vi.fn(),
  setInputSchemaDefault: vi.fn(),
  deleteInputSchema: vi.fn(),
  createInputMapper: vi.fn(),
  updateInputMapper: vi.fn(),
  setInputMapperActive: vi.fn(),
  deleteInputMapper: vi.fn(),
  createCompatibility: vi.fn(),
  updateCompatibility: vi.fn(),
  setCompatibilityDefaultRoute: vi.fn(),
  deactivateCompatibility: vi.fn(),
  deleteCompatibility: vi.fn(),
  createExperiment: vi.fn(),
  updateExperiment: vi.fn(),
  updateExperimentStatus: vi.fn(),
  deleteExperiment: vi.fn(),
  createExperimentScope: vi.fn(),
  updateExperimentScope: vi.fn(),
  deleteExperimentScope: vi.fn(),
  createExperimentVariant: vi.fn(),
  updateExperimentVariant: vi.fn(),
  deleteExperimentVariant: vi.fn(),
}))

vi.mock('~/components/ui/toast/use-toast', () => ({
  toast: toastMock,
}))

describe('workflow routing modal forms', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    toastMock.mockClear()
    routerPushMock.mockClear()
    routeLeaveMock.mockClear()
    for (const mock of Object.values(routingMock)) mock.mockReset()

    routingMock.fetchInputSchemas.mockResolvedValue([{
      id: 11,
      workflow_id: 42,
      code: 'public',
      version_number: 1,
      schema_json: { type: 'object', properties: { email: { type: 'string' } } },
      status: 'active',
      is_default: true,
    }])
    routingMock.fetchInputMappers.mockResolvedValue([{
      id: 21,
      workflow_id: 42,
      name: 'Map public payload',
      mapper_type: 'json',
      rules: { copy_all: true },
      is_active: true,
    }])
    routingMock.fetchCompatibilities.mockResolvedValue([{
      id: 31,
      workflow_version_id: 101,
      workflow_input_schema_id: 11,
      workflow_input_mapper_id: 21,
      compatibility_type: 'native',
      default_values: {},
      is_active: true,
      is_default_route: true,
    }])
    routingMock.fetchExperiments.mockResolvedValue([{
      id: 41,
      workflow_id: 42,
      name: 'Canary route',
      description: undefined,
      experiment_type: 'canary',
      status: 'draft',
      started_at: undefined,
      ended_at: undefined,
    }])
    routingMock.fetchExperimentScopes.mockResolvedValue([{
      id: 51,
      workflow_experiment_id: 41,
      workflow_input_schema_id: 11,
      traffic_conditions: {},
      conditions_hash: 'hash',
      traffic_percent: 50,
      fallback_policy: 'default_route',
      fallback_workflow_version_id: undefined,
    }])
    routingMock.fetchExperimentVariants.mockResolvedValue([{
      id: 61,
      workflow_experiment_scope_id: 51,
      workflow_version_id: 101,
      traffic_weight: 50,
      is_control_group: true,
      is_active: true,
    }])
    routingMock.deleteInputSchema.mockResolvedValue(undefined)
    routingMock.setInputSchemaDefault.mockResolvedValue(undefined)
    routingMock.updateInputSchemaStatus.mockResolvedValue(undefined)
    routingMock.setInputMapperActive.mockResolvedValue(undefined)
    routingMock.setCompatibilityDefaultRoute.mockResolvedValue(undefined)
    routingMock.deactivateCompatibility.mockResolvedValue(undefined)
    routingMock.updateExperimentStatus.mockResolvedValue(undefined)

    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('reactive', reactive)
    vi.stubGlobal('onMounted', (callback: () => void) => callback())
    vi.stubGlobal('onBeforeRouteLeave', routeLeaveMock)
    vi.stubGlobal('getErrorMessage', getErrorMessage)
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42' },
    }))
    vi.stubGlobal('useRouter', () => ({
      push: routerPushMock,
    }))
    vi.stubGlobal('useI18n', () => ({
      t: translate,
    }))
    vi.stubGlobal('useWorkflows', () => ({
      fetchWorkflow: vi.fn().mockResolvedValue({ id: 42, name: 'Process' }),
    }))
    vi.stubGlobal('useVersions', () => ({
      fetchVersionSummaries: vi.fn().mockResolvedValue([{
        id: 101,
        workflow_id: 42,
        name: 'Current',
        version_number: 3,
        is_valid: true,
        is_active: true,
        traffic_weight: 100,
        is_control_group: true,
        run_count: 0,
        created_at: '',
      }]),
    }))
    vi.stubGlobal('useWorkflowRouting', () => routingMock)
    Object.defineProperty(window, 'confirm', {
      value: vi.fn(() => true),
      configurable: true,
    })
  })

  it('routes input schema create and edit to pages, lists fields, and confirms deletes in a dialog', async () => {
    const wrapper = mount(WorkflowRoutingPage, {
      global: {
        stubs: pageStubs(),
      },
    })
    await flushPromises()

    expect(wrapper.text()).not.toContain('Add or edit scope and variant')
    expect(textCount(wrapper.text(), 'Create input schema')).toBe(1)
    expect(textCount(wrapper.text(), 'Create mapper')).toBe(1)
    expect(textCount(wrapper.text(), 'Create compatibility')).toBe(1)
    expect(textCount(wrapper.text(), 'Create experiment')).toBe(1)
    expect(wrapper.text()).toContain('email')
    expect(wrapper.text()).not.toContain('"properties"')
    expect(wrapper.find('[data-testid="routing-input-schema-fields-11"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="routing-input-schema-field-list-11"]').exists()).toBe(true)

    await wrapper.get('[data-testid="open-create-schema-modal"]').trigger('click')

    expect(routerPushMock).toHaveBeenCalledWith('/org/7/workflows/42/routing/input-schemas/new')

    await wrapper.get('[data-testid="edit-schema-11"]').trigger('click')

    expect(routerPushMock).toHaveBeenCalledWith('/org/7/workflows/42/routing/input-schemas/11/edit')

    await wrapper.get('[data-testid="delete-schema-11"]').trigger('click')

    expect(routingMock.deleteInputSchema).not.toHaveBeenCalled()
    expect(wrapper.get('[data-testid="routing-delete-dialog"]').text()).toContain('Delete or archive input schema public v1?')

    await wrapper.get('[data-testid="confirm-routing-delete"]').trigger('click')
    await flushPromises()

    expect(routingMock.deleteInputSchema).toHaveBeenCalledWith(11)
  })

  it('does not register route leave guard now that local routing changes are not tracked', async () => {
    mount(WorkflowRoutingPage, {
      global: {
        stubs: pageStubs(),
      },
    })
    await flushPromises()

    expect(routeLeaveMock).not.toHaveBeenCalled()
  })

  it('shows field labels above every routing form control', () => {
    const global = { stubs: pageStubs() }

    expectDialogLabels(mount(RoutingInputMapperDialog, {
      props: {
        open: true,
        form: {
          name: 'Map public payload',
          mapperType: 'json',
          rulesJson: '{}',
        },
        saving: false,
        editingId: null,
      },
      global,
    }), [
      'Mapper name',
      'Mapper type',
      'Mapper rules JSON',
    ])

    expectDialogLabels(mount(RoutingCompatibilityDialog, {
      props: {
        open: true,
        form: {
          versionId: '',
          inputSchemaId: '',
          mapperId: '',
          compatibilityType: 'native',
          isDefaultRoute: false,
          defaultValuesJson: '{}',
        },
        activeVersions: [{
          id: 101,
          workflow_id: 42,
          name: 'Current',
          version_number: 3,
          is_valid: true,
          is_active: true,
          traffic_weight: 100,
          is_control_group: true,
          run_count: 0,
          created_at: '',
        }],
        activeSchemas: [{
          id: 11,
          workflow_id: 42,
          code: 'public',
          version_number: 1,
          schema_json: {},
          status: 'active',
          is_default: true,
        }],
        mappers: [{
          id: 21,
          workflow_id: 42,
          name: 'Map public payload',
          mapper_type: 'json',
          rules: {},
          is_active: true,
        }],
        saving: false,
        editingId: null,
        versionLabel: (versionId: number) => `Version ${versionId}`,
        schemaLabel: (schemaId: number) => `Schema ${schemaId}`,
      },
      global,
    }), [
      'Workflow version',
      'Input schema',
      'Input mapper',
      'Compatibility type',
      'Default route',
      'Default values JSON',
    ])

    expectDialogLabels(mount(RoutingExperimentDialog, {
      props: {
        open: true,
        form: {
          name: 'Canary route',
          experimentType: 'canary',
          status: 'draft',
        },
        saving: false,
        editingId: null,
      },
      global,
    }), [
      'Experiment name',
      'Experiment type',
      'Status',
    ])

    expectDialogLabels(mount(RoutingScopeDialog, {
      props: {
        open: true,
        form: {
          experimentId: '',
          inputSchemaId: '',
          trafficPercent: 100,
          fallbackPolicy: 'default_route',
          fallbackVersionId: '',
          trafficConditionsJson: '{}',
        },
        experiments: [{
          id: 41,
          workflow_id: 42,
          name: 'Canary route',
          experiment_type: 'canary',
          status: 'draft',
        }],
        activeSchemas: [{
          id: 11,
          workflow_id: 42,
          code: 'public',
          version_number: 1,
          schema_json: {},
          status: 'active',
          is_default: true,
        }],
        saving: false,
        editingId: null,
        schemaLabel: (schemaId: number) => `Schema ${schemaId}`,
      },
      global,
    }), [
      'Experiment',
      'Input schema',
      'Traffic percent',
      'Traffic conditions JSON',
    ])

    expectDialogLabels(mount(RoutingVariantDialog, {
      props: {
        open: true,
        form: {
          scopeId: '',
          versionId: '',
          trafficWeight: 100,
          isControlGroup: false,
          isActive: true,
        },
        experiments: [{
          id: 41,
          workflow_id: 42,
          name: 'Canary route',
          experiment_type: 'canary',
          status: 'draft',
        }],
        scopesByExperiment: {
          41: [{
            id: 51,
            workflow_experiment_id: 41,
            workflow_input_schema_id: 11,
            traffic_conditions: {},
            conditions_hash: 'hash',
            traffic_percent: 50,
            fallback_policy: 'default_route',
            fallback_workflow_version_id: undefined,
          }],
        },
        activeVersions: [{
          id: 101,
          workflow_id: 42,
          name: 'Current',
          version_number: 3,
          is_valid: true,
          is_active: true,
          traffic_weight: 100,
          is_control_group: true,
          run_count: 0,
          created_at: '',
        }],
        saving: false,
        editingId: null,
        versionLabel: (versionId: number) => `Version ${versionId}`,
        schemaLabel: (schemaId: number) => `Schema ${schemaId}`,
      },
      global,
    }), [
      'Scope',
      'Workflow version',
      'Traffic weight',
      'Control group',
      'Active variant',
    ])
  })
})

function expectDialogLabels(wrapper: { text: () => string }, labels: string[]) {
  const text = wrapper.text()
  for (const label of labels) {
    expect(text).toContain(label)
  }
}

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
        return () => h('button', { ...attrs, disabled: props.disabled }, slots.default?.())
      },
    }),
    Card: passthrough,
    CardContent: passthrough,
    CardDescription: { template: '<p><slot /></p>' },
    CardHeader: passthrough,
    CardTitle: { template: '<h2><slot /></h2>' },
    Dialog: defineComponent({
      props: {
        open: { type: Boolean, default: false },
      },
      setup(props, { slots }) {
        return () => props.open ? h('div', slots.default?.()) : null
      },
    }),
    DialogContent: passthrough,
    DialogDescription: { template: '<p><slot /></p>' },
    DialogFooter: passthrough,
    DialogHeader: passthrough,
    DialogTitle: { template: '<h2><slot /></h2>' },
    Input: defineComponent({
      props: {
        modelValue: { type: [String, Number], default: '' },
        disabled: { type: Boolean, default: false },
      },
      emits: ['update:modelValue'],
      setup(props, { attrs, emit }) {
        return () => h('input', {
          ...attrs,
          disabled: props.disabled,
          value: props.modelValue,
          onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
        })
      },
    }),
  }
}

function translate(key: string, params?: Record<string, unknown>) {
  const translations: Record<string, string> = {
    'common.back': 'Back',
    'common.cancel': 'Cancel',
    'common.delete': 'Delete',
    'common.edit': 'Edit',
    'common.loading': 'Loading',
    'common.save': 'Save',
    'destructive.cancel': 'Cancel',
    'destructive.delete': 'Delete',
    'error.server': 'Server error',
    'status.active': 'active',
    'status.inactive': 'inactive',
    'workflowRouting.activate': 'Activate',
    'workflowRouting.addOrEditScopeVariant': 'Add or edit scope and variant',
    'workflowRouting.compatibilities': 'Schema compatibilities',
    'workflowRouting.compatibilitiesDescription': 'Routes from input schemas to executable workflow versions.',
    'workflowRouting.confirmDeleteSchema': `Delete or archive input schema ${params?.name}?`,
    'workflowRouting.createCompatibility': 'Create compatibility',
    'workflowRouting.createExperiment': 'Create experiment',
    'workflowRouting.createInputSchema': 'Create input schema',
    'workflowRouting.createMapper': 'Create mapper',
    'workflowRouting.createScope': 'Create scope',
    'workflowRouting.createVariant': 'Create variant',
    'workflowRouting.deactivate': 'Deactivate',
    'workflowRouting.default': 'default',
    'workflowRouting.defaultRoute': 'default route',
    'workflowRouting.deleteScope': 'Delete scope',
    'workflowRouting.deletedInputSchema': 'Input schema deleted or archived',
    'workflowRouting.discardChanges': 'Discard changes',
    'workflowRouting.editCompatibility': 'Edit compatibility',
    'workflowRouting.editExperiment': 'Edit experiment',
    'workflowRouting.editInputSchema': 'Edit input schema',
    'workflowRouting.editMapper': 'Edit mapper',
    'workflowRouting.editScope': 'Edit scope',
    'workflowRouting.errorCreateInputSchema': 'Failed to create input schema',
    'workflowRouting.experiments': 'Experiments',
    'workflowRouting.experimentsDescription': 'Active experiments override default routes using deterministic buckets.',
    'workflowRouting.fieldActiveVariant': 'Active variant',
    'workflowRouting.fieldCompatibilityType': 'Compatibility type',
    'workflowRouting.fieldControlGroup': 'Control group',
    'workflowRouting.fieldDefaultRoute': 'Default route',
    'workflowRouting.fieldDefaultSchema': 'Default schema',
    'workflowRouting.fieldDefaultValuesJson': 'Default values JSON',
    'workflowRouting.fieldEnabled': 'Enabled',
    'workflowRouting.fieldExperiment': 'Experiment',
    'workflowRouting.fieldExperimentName': 'Experiment name',
    'workflowRouting.fieldExperimentType': 'Experiment type',
    'workflowRouting.fieldInputMapper': 'Input mapper',
    'workflowRouting.fieldInputSchema': 'Input schema',
    'workflowRouting.fieldMapperName': 'Mapper name',
    'workflowRouting.fieldMapperRulesJson': 'Mapper rules JSON',
    'workflowRouting.fieldMapperType': 'Mapper type',
    'workflowRouting.fieldSchemaCode': 'Schema code',
    'workflowRouting.fieldSchemaJson': 'Schema JSON',
    'workflowRouting.fieldScope': 'Scope',
    'workflowRouting.fieldStatus': 'Status',
    'workflowRouting.fieldTrafficConditionsJson': 'Traffic conditions JSON',
    'workflowRouting.fieldTrafficPercent': 'Traffic percent',
    'workflowRouting.fieldTrafficWeight': 'Traffic weight',
    'workflowRouting.fieldWorkflowVersion': 'Workflow version',
    'workflowRouting.inputMappers': 'Input mappers',
    'workflowRouting.inputMappersDescription': 'Payload transforms used by compatibility routes.',
    'workflowRouting.inputSchemas': 'Input schemas',
    'workflowRouting.inputSchemasDescription': 'Versioned public request contracts used by message routing.',
    'workflowRouting.makeDefaultRoute': 'Make default route',
    'workflowRouting.noCompatibilities': 'No schema compatibilities yet.',
    'workflowRouting.noExperiments': 'No experiments yet.',
    'workflowRouting.noInputMappers': 'No input mappers yet.',
    'workflowRouting.noInputSchemas': 'No input schemas yet.',
    'workflowRouting.pause': 'Pause',
    'workflowRouting.placeholderCompatibilityType': 'native',
    'workflowRouting.placeholderExperimentName': 'experiment name',
    'workflowRouting.placeholderMapperName': 'mapper name',
    'workflowRouting.placeholderMapperType': 'internal',
    'workflowRouting.placeholderSchemaCode': 'schema code',
    'workflowRouting.routes': 'Routes',
    'workflowRouting.schemas': 'Schemas',
    'workflowRouting.scopeNumber': `scope #${params?.id}`,
    'workflowRouting.setDefault': 'Set default',
    'workflowRouting.statusActive': 'Active',
    'workflowRouting.statusDraft': 'Draft',
    'workflowRouting.statusPaused': 'Paused',
    'workflowRouting.subtitle': 'Schemas, route compatibility, mappers, and experiments for runtime selection.',
    'workflowRouting.title': 'Workflow routing',
    'workflowRouting.to': 'to',
    'workflowRouting.typeCanary': 'Canary',
    'workflowRouting.typeRollout': 'Rollout',
    'workflowRouting.typeShadow': 'Shadow',
    'workflowRouting.typeSplit': 'Split',
    'workflowRouting.unsavedChanges': 'Unsaved routing changes',
    'workflowRouting.unsavedChangesDescription': `Local changes in ${params?.forms} are not saved yet.`,
    'workflowRouting.updateCompatibility': 'Update compatibility',
    'workflowRouting.updateExperiment': 'Update experiment',
    'workflowRouting.updateInputSchema': 'Update schema',
    'workflowRouting.updateMapper': 'Update mapper',
    'workflowRouting.updateScope': 'Update scope',
    'workflowRouting.updateVariant': 'Update variant',
    'workflowRouting.variant': 'variant',
  }
  return translations[key] ?? key
}

function textCount(text: string, needle: string) {
  return text.split(needle).length - 1
}

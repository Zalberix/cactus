import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, h, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import VersionEditorPage from '../app/pages/org/[orgId]/workflows/[workflowId]/versions/[versionId]/edit.vue'
import { getErrorMessage } from '../app/utils/errors'

const toastMock = vi.hoisted(() => vi.fn())

vi.mock('~/components/ui/toast/use-toast', () => ({
  toast: toastMock,
}))

describe('workflow step delete confirmation', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    toastMock.mockClear()

    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('onMounted', (callback: () => void) => callback())
    vi.stubGlobal('useI18n', () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'editor.versionNumber') return `Version ${params?.number ?? ''}`
        if (key === 'destructive.deleteStep.body') return `Delete ${params?.name ?? ''}?`
        return key
      },
    }))
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42', versionId: '10' },
    }))
    vi.stubGlobal('useRouter', () => ({ replace: vi.fn() }))
    vi.stubGlobal('getErrorMessage', getErrorMessage)
    vi.stubGlobal('useWorkflowRouting', () => ({
      fetchCompatibilities: vi.fn(),
      fetchInputSchemas: vi.fn(),
    }))
  })

  it('requires confirmation before deleting a step', async () => {
    const removeStep = vi.fn()

    vi.stubGlobal('useWorkflows', () => ({
      fetchWorkflow: vi.fn().mockResolvedValue({ id: 42, name: 'Process' }),
      fetchWorkflowInputSchema: vi.fn(),
      upsertWorkflowInputSchemaField: vi.fn(),
      deleteWorkflowInputSchemaField: vi.fn(),
      workflowVersionEditorPath: vi.fn(),
    }))
    vi.stubGlobal('useVersions', () => ({
      fetchVersionSummaries: vi.fn().mockResolvedValue([{
        id: 10,
        workflow_id: 42,
        name: 'Current',
        version_number: 1,
        is_valid: false,
        is_active: false,
        traffic_weight: 100,
        is_control_group: false,
        run_count: 0,
        created_at: '',
      }]),
      copyVersion: vi.fn(),
      updateVersionName: vi.fn(),
      activateVersion: vi.fn(),
      deactivateVersion: vi.fn(),
    }))
    vi.stubGlobal('useDagEditor', () => ({
      nodes: ref([
        {
          id: '12',
          data: {
            label: 'Send email',
            controlKind: undefined,
          },
        },
      ]),
      edges: ref([]),
      selectedEdgeId: ref(null),
      selectedNodeId: ref('12'),
      isDirty: ref(false),
      validationErrors: ref([]),
      loadSteps: vi.fn(),
      saveVersion: vi.fn(),
      connectSteps: vi.fn(),
      onNodeDragStop: vi.fn(),
      selectNode: vi.fn(),
      removeEdge: vi.fn(),
      addStep: vi.fn(),
      removeStep,
      renameStepOnServer: vi.fn(),
      updateTaskSettingsOnServer: vi.fn(),
      updateStepOnServer: vi.fn(),
      updateTaskInputMappingOnServer: vi.fn(),
      updateNodeData: vi.fn(),
    }))
    vi.stubGlobal('useNodeEditor', () => ({
      isOpen: ref(false),
      editingNodeId: ref(null),
      open: vi.fn(),
      close: vi.fn(),
    }))

    const wrapper = mount(VersionEditorPage, {
      global: {
        stubs: pageStubs(),
      },
    })
    await flushPromises()

    await wrapper.get('[data-testid="delete-selected-event"]').trigger('click')

    expect(removeStep).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('destructive.deleteStep.title')
    expect(wrapper.text()).toContain('Delete Send email?')

    await wrapper.get('[data-testid="cancel-step-delete"]').trigger('click')

    expect(removeStep).not.toHaveBeenCalled()

    await wrapper.get('[data-testid="delete-node-event"]').trigger('click')
    await wrapper.get('[data-testid="confirm-step-delete"]').trigger('click')

    expect(removeStep).toHaveBeenCalledWith('12')
  })
})

function pageStubs() {
  const passthrough = { template: '<div><slot /></div>' }
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
    DagCanvas: defineComponent({
      emits: ['deleteSelected', 'deleteNode'],
      setup(_, { emit }) {
        return () => h('div', [
          h('button', {
            'data-testid': 'delete-selected-event',
            onClick: () => emit('deleteSelected'),
          }),
          h('button', {
            'data-testid': 'delete-node-event',
            onClick: () => emit('deleteNode', '12'),
          }),
        ])
      },
    }),
    Dialog: defineComponent({
      props: {
        open: { type: Boolean, default: false },
      },
      emits: ['update:open'],
      setup(props, { slots }) {
        return () => props.open ? h('div', slots.default?.()) : null
      },
    }),
    DialogContent: passthrough,
    DialogDescription: passthrough,
    DialogFooter: passthrough,
    DialogHeader: passthrough,
    DialogTitle: { template: '<h2><slot /></h2>' },
    EmptyState: { template: '<div />' },
    Input: defineComponent({
      props: {
        modelValue: { type: String, default: '' },
        readonly: { type: Boolean, default: false },
      },
      emits: ['update:modelValue', 'blur', 'keydown'],
      setup(props, { attrs, emit }) {
        return () => h('input', {
          ...attrs,
          readonly: props.readonly,
          value: props.modelValue,
          onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
          onBlur: (event: Event) => emit('blur', event),
          onKeydown: (event: KeyboardEvent) => emit('keydown', event),
        })
      },
    }),
    NodeEditor: { template: '<div />' },
    StepRenameDialog: { template: '<div />' },
    StepSchemaChoiceDialog: { template: '<div />' },
    StepToolbar: { template: '<div />' },
    WorkflowSchemaDialog: { template: '<div />' },
  }
}

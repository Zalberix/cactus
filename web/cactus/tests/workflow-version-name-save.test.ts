import { flushPromises, mount } from '@vue/test-utils'
import { computed, defineComponent, h, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import VersionEditorPage from '../app/pages/org/[orgId]/workflows/[workflowId]/versions/[versionId]/edit.vue'
import { getErrorMessage } from '../app/utils/errors'

const toastMock = vi.hoisted(() => vi.fn())

vi.mock('~/components/ui/toast/use-toast', () => ({
  toast: toastMock,
}))

describe('workflow version name save', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    toastMock.mockClear()

    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('onMounted', (callback: () => void) => callback())
    vi.stubGlobal('useI18n', () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        key === 'editor.versionNumber' ? `Version ${params?.number ?? ''}` : key,
    }))
    vi.stubGlobal('useRoute', () => ({
      params: { orgId: '7', workflowId: '42', versionId: '10' },
    }))
    vi.stubGlobal('useRouter', () => ({ replace: vi.fn() }))
    vi.stubGlobal('getErrorMessage', getErrorMessage)
  })

  it('shows the API error only after saving a duplicate name', async () => {
    const updateVersionName = vi.fn().mockRejectedValue(new Error('Такое имя версии уже существует'))

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
      updateVersionName,
      activateVersion: vi.fn(),
      deactivateVersion: vi.fn(),
    }))
    vi.stubGlobal('useDagEditor', () => ({
      nodes: ref([]),
      edges: ref([]),
      selectedEdgeId: ref(null),
      selectedNodeId: ref(null),
      isDirty: ref(false),
      validationErrors: ref([]),
      loadSteps: vi.fn(),
      saveVersion: vi.fn(),
      connectSteps: vi.fn(),
      onNodeDragStop: vi.fn(),
      selectNode: vi.fn(),
      removeEdge: vi.fn(),
      addStep: vi.fn(),
      removeStep: vi.fn(),
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

    await wrapper.get('input').setValue('Release 42')
    expect(toastMock).not.toHaveBeenCalledWith(expect.objectContaining({
      title: 'Такое имя версии уже существует',
    }))

    await wrapper.get('input').trigger('blur')
    await flushPromises()

    expect(updateVersionName).toHaveBeenCalledWith(10, 'Release 42')
    expect(toastMock).toHaveBeenCalledWith({
      title: 'Такое имя версии уже существует',
      variant: 'destructive',
    })
  })

  it('notifies breadcrumbs after the version name is saved', async () => {
    const updateVersionName = vi.fn().mockResolvedValue({
      id: 10,
      workflow_id: 42,
      name: 'Release 42',
      version_number: 1,
      is_valid: false,
      is_active: false,
      created_at: '',
    })
    const fetchVersionSummaries = vi.fn().mockResolvedValue([{
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
    }])
    const events: CustomEvent[] = []
    const onNameUpdated = (event: Event) => events.push(event as CustomEvent)
    window.addEventListener('cactus:workflow-version-name-updated', onNameUpdated)

    vi.stubGlobal('useWorkflows', () => ({
      fetchWorkflow: vi.fn().mockResolvedValue({ id: 42, name: 'Process' }),
      fetchWorkflowInputSchema: vi.fn(),
      upsertWorkflowInputSchemaField: vi.fn(),
      deleteWorkflowInputSchemaField: vi.fn(),
      workflowVersionEditorPath: vi.fn(),
    }))
    vi.stubGlobal('useVersions', () => ({
      fetchVersionSummaries,
      copyVersion: vi.fn(),
      updateVersionName,
      activateVersion: vi.fn(),
      deactivateVersion: vi.fn(),
    }))
    vi.stubGlobal('useDagEditor', () => ({
      nodes: ref([]),
      edges: ref([]),
      selectedEdgeId: ref(null),
      selectedNodeId: ref(null),
      isDirty: ref(false),
      validationErrors: ref([]),
      loadSteps: vi.fn(),
      saveVersion: vi.fn(),
      connectSteps: vi.fn(),
      onNodeDragStop: vi.fn(),
      selectNode: vi.fn(),
      removeEdge: vi.fn(),
      addStep: vi.fn(),
      removeStep: vi.fn(),
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

    try {
      const wrapper = mount(VersionEditorPage, {
        global: {
          stubs: pageStubs(),
        },
      })
      await flushPromises()

      await wrapper.get('input').setValue('Release 42')
      await wrapper.get('input').trigger('blur')
      await flushPromises()

      expect(updateVersionName).toHaveBeenCalledWith(10, 'Release 42')
      expect(events).toHaveLength(1)
      expect(events[0].detail).toEqual({
        versionId: 10,
        name: 'Release 42',
        versionNumber: 1,
      })
    }
    finally {
      window.removeEventListener('cactus:workflow-version-name-updated', onNameUpdated)
    }
  })
})

function pageStubs() {
  const passthrough = { template: '<div><slot /></div>' }
  return {
    Badge: { template: '<span><slot /></span>' },
    Button: defineComponent({
      setup(_, { attrs, slots }) {
        return () => h('button', attrs, slots.default?.())
      },
    }),
    Dialog: passthrough,
    DialogContent: passthrough,
    DialogDescription: passthrough,
    DialogFooter: passthrough,
    DialogHeader: passthrough,
    DialogTitle: passthrough,
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
    DagCanvas: { template: '<div />' },
    EmptyState: { template: '<div />' },
    NodeEditor: { template: '<div />' },
    StepRenameDialog: { template: '<div />' },
    StepSchemaChoiceDialog: { template: '<div />' },
    StepToolbar: { template: '<div />' },
    WorkflowSchemaDialog: { template: '<div />' },
  }
}

import { flushPromises, mount } from '@vue/test-utils'
import { computed, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import InputPanel from '../app/components/dag/node-editor/InputPanel.vue'

describe('InputPanel workflow input actions', () => {
  const fetchWorkflowInputSchema = vi.fn()
  const upsertWorkflowInputSchemaField = vi.fn()
  const deleteWorkflowInputSchemaField = vi.fn()
  const fetchInputSchema = vi.fn()
  const updateInputSchema = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
    vi.stubGlobal('useWorkflows', () => ({
      fetchWorkflowInputSchema,
      upsertWorkflowInputSchemaField,
      deleteWorkflowInputSchemaField,
    }))
    vi.stubGlobal('useWorkflowRouting', () => ({
      fetchInputSchema,
      updateInputSchema,
    }))

    fetchWorkflowInputSchema.mockResolvedValue({
      type: 'object',
      properties: {
        email: { type: 'string', required: true },
      },
    })
    upsertWorkflowInputSchemaField.mockResolvedValue({})
    deleteWorkflowInputSchemaField.mockResolvedValue({})
    fetchInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'v2',
      version_number: 2,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
        },
      },
      status: 'draft',
      is_default: false,
    })
    updateInputSchema.mockResolvedValue({
      id: 11,
      workflow_id: 42,
      code: 'v2',
      version_number: 2,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
          phone: { type: 'string' },
        },
      },
      status: 'draft',
      is_default: false,
    })
  })

  it('creates workflow inputs in the linked native input schema', async () => {
    const wrapper = mount(InputPanel, {
      props: {
        stepId: '10',
        workflowId: 42,
        versionId: 8,
        workflowInputSchemaId: 11,
        allNodes: [],
        allEdges: [],
      },
      global: {
        stubs: {
          ScrollArea: { template: '<div><slot /></div>' },
          Separator: { template: '<hr>' },
          SchemaTree: { template: '<div />' },
          WorkflowInputsPanel: {
            props: ['inputs'],
            emits: ['deleteInput'],
            template: '<button data-testid="delete-email" @click="$emit(\'deleteInput\', inputs[0])">delete</button>',
          },
        },
      },
    })

    await flushPromises()
    const onCreated = vi.fn()

    await (wrapper.vm as unknown as {
      createWorkflowInputFromField: (field: string, property: Record<string, unknown>, onCreated: (expression: string) => void) => Promise<void>
    }).createWorkflowInputFromField('phone', { type: 'string' }, onCreated)
    await flushPromises()

    expect(fetchInputSchema).toHaveBeenCalledWith(11)
    expect(updateInputSchema).toHaveBeenCalledWith(11, {
      code: 'v2',
      version_number: 2,
      schema_json: {
        type: 'object',
        properties: {
          email: { type: 'string', required: true },
          phone: { type: 'string' },
        },
      },
    })
    expect(upsertWorkflowInputSchemaField).not.toHaveBeenCalled()
    expect(onCreated).toHaveBeenCalledWith('$.message.value.phone')
    expect(wrapper.emitted('workflowInputsChanged')).toEqual([[]])
  })

  it('deletes workflow inputs from the linked native input schema', async () => {
    const wrapper = mount(InputPanel, {
      props: {
        stepId: '10',
        workflowId: 42,
        versionId: 8,
        workflowInputSchemaId: 11,
        allNodes: [],
        allEdges: [],
      },
      global: {
        stubs: {
          ScrollArea: { template: '<div><slot /></div>' },
          Separator: { template: '<hr>' },
          SchemaTree: { template: '<div />' },
          WorkflowInputsPanel: {
            props: ['inputs'],
            emits: ['deleteInput'],
            template: '<button data-testid="delete-email" @click="$emit(\'deleteInput\', inputs[0])">delete</button>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('[data-testid="delete-email"]').trigger('click')
    await wrapper.get('[data-testid="confirm-workflow-input-delete"]').trigger('click')
    await flushPromises()

    expect(updateInputSchema).toHaveBeenCalledWith(11, {
      code: 'v2',
      version_number: 2,
      schema_json: {
        type: 'object',
        properties: {},
      },
    })
    expect(deleteWorkflowInputSchemaField).not.toHaveBeenCalled()
    expect(wrapper.emitted('workflowInputsChanged')).toEqual([[]])
  })
})

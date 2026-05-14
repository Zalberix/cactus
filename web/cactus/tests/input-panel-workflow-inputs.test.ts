import { flushPromises, mount } from '@vue/test-utils'
import { computed, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import InputPanel from '../app/components/dag/node-editor/InputPanel.vue'

describe('InputPanel workflow input actions', () => {
  const fetchWorkflowInputSchema = vi.fn()
  const upsertWorkflowInputSchemaField = vi.fn()
  const deleteWorkflowInputSchemaField = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
    vi.stubGlobal('useWorkflows', () => ({
      fetchWorkflowInputSchema,
      upsertWorkflowInputSchemaField,
      deleteWorkflowInputSchemaField,
    }))

    fetchWorkflowInputSchema.mockResolvedValue({
      type: 'object',
      properties: {
        email: { type: 'string', required: true },
      },
    })
    upsertWorkflowInputSchemaField.mockResolvedValue({})
    deleteWorkflowInputSchemaField.mockResolvedValue({})
  })

  it('emits workflowInputsChanged after deleting a workflow input', async () => {
    const wrapper = mount(InputPanel, {
      props: {
        stepId: '10',
        workflowId: 42,
        versionId: 8,
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

    expect(deleteWorkflowInputSchemaField).toHaveBeenCalledWith(42, 'email')
    expect(wrapper.emitted('workflowInputsChanged')).toEqual([[]])
  })
})

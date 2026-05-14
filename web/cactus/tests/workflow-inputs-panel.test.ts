import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { workflowInputFieldsFromSchema, workflowInputPath } from '../app/components/dag/node-editor/workflow-input-utils'
import WorkflowInputsPanel from '../app/components/dag/node-editor/WorkflowInputsPanel.vue'

describe('workflow inputs panel', () => {
  beforeEach(() => {
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  })

  it('builds message value path for workflow input', () => {
    expect(workflowInputPath('email')).toBe('$.message.value.email')
  })

  it('builds sorted workflow input fields from schema properties', () => {
    expect(workflowInputFieldsFromSchema({
      type: 'object',
      properties: {
        count: { type: 'integer' },
        email: { type: 'string', required: true, description: 'Recipient address' },
      },
    })).toEqual([
      {
        name: 'count',
        type: 'integer',
        required: false,
      },
      {
        name: 'email',
        type: 'string',
        required: true,
        description: 'Recipient address',
      },
    ])
  })

  it('emits edit and delete actions from the overflow menu without inserting the input expression', async () => {
    const input = {
      name: 'email',
      type: 'string' as const,
      required: true,
      description: 'Recipient address',
    }
    const wrapper = mount(WorkflowInputsPanel, {
      props: { inputs: [input] },
      global: {
        stubs: {
          Badge: { template: '<span><slot /></span>' },
          DropdownMenu: { template: '<div><slot /></div>' },
          DropdownMenuTrigger: { template: '<div><slot /></div>' },
          DropdownMenuContent: { template: '<div><slot /></div>' },
          DropdownMenuItem: { template: '<button type="button"><slot /></button>' },
        },
      },
    })

    expect(wrapper.get('[data-testid="workflow-input-actions-email"]').text()).toContain('...')
    await wrapper.get('[data-testid="edit-input-email"]').trigger('click')
    await wrapper.get('[data-testid="delete-input-email"]').trigger('click')

    expect(wrapper.emitted('editInput')).toEqual([[input]])
    expect(wrapper.emitted('deleteInput')).toEqual([[input]])
    expect(wrapper.emitted('insertExpression')).toBeUndefined()
  })
})

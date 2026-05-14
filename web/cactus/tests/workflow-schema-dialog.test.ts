import { mount } from '@vue/test-utils'
import { computed } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import WorkflowSchemaDialog from '../app/components/dag/WorkflowSchemaDialog.vue'

describe('workflow schema dialog', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('useI18n', () => ({
      t: (key: string) => ({
        'editor.inputSchema': 'Input schema',
        'editor.inputSchemaDescription': 'Workflow input schema',
        'workflowInputs.name': 'Name',
        'workflowInputs.type': 'Type',
        'workflowInputs.required': 'Required',
        'workflowInputs.description': 'Description',
        'workflowInputs.actions': 'Actions',
      }[key] ?? key),
    }))
  })

  it('renders input schema fields as a table', () => {
    const wrapper = mount(WorkflowSchemaDialog, {
      props: {
        open: true,
        schema: {
          type: 'object',
          properties: {
            email: { type: 'string', required: true, description: 'Recipient address' },
          },
        },
      },
      global: {
        stubs: {
          Dialog: { template: '<div><slot /></div>' },
          DialogContent: { template: '<section><slot /></section>' },
          DialogHeader: { template: '<header><slot /></header>' },
          DialogTitle: { template: '<h2><slot /></h2>' },
          DialogDescription: { template: '<p><slot /></p>' },
        },
      },
    })

    expect(wrapper.find('[data-testid="workflow-input-schema-table"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="workflow-input-row-email"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="workflow-input-type-email"]').text()).toBe('string')
    expect(wrapper.get('[data-testid="workflow-input-required-email"]').text()).toBe('Required')
    expect(wrapper.text()).toContain('Recipient address')
  })
})

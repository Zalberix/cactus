import { describe, expect, it } from 'vitest'
import { workflowInputFieldsFromSchema, workflowInputPath } from '../app/components/dag/node-editor/workflow-input-utils'

describe('workflow inputs panel', () => {
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
})

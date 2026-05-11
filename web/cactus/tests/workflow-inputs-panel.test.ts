import { describe, expect, it } from 'vitest'
import { workflowInputPath } from '../app/components/dag/node-editor/workflow-input-utils'

describe('workflow inputs panel', () => {
  it('builds message value path for workflow input', () => {
    expect(workflowInputPath('email')).toBe('$.message.value.email')
  })
})

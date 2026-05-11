import { describe, expect, it } from 'vitest'
import { schemaPropertiesForForm } from '../app/components/forms/dynamic-settings-utils'

describe('dynamic settings form schema parsing', () => {
  it('marks fields required from root JSON Schema required array', () => {
    const schema = {
      type: 'object',
      required: ['email'],
      properties: {
        email: { type: 'string' },
        subject: { type: 'string' },
      },
    }

    const fields = schemaPropertiesForForm(schema)

    expect(fields.find(f => f.key === 'email')?.isRequired).toBe(true)
    expect(fields.find(f => f.key === 'subject')?.isRequired).toBe(false)
  })
})

import { describe, expect, it } from 'vitest'
import { schemaPropertiesForForm, validateSettingsData } from '../app/components/forms/dynamic-settings-utils'

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

describe('validateSettingsData', () => {
  it('rejects missing required settings', () => {
    expect(validateSettingsData({
      type: 'object',
      properties: { host: { type: 'string', required: true } },
    }, {})).toContain('host')
  })

  it('accepts valid required settings', () => {
    expect(validateSettingsData({
      type: 'object',
      properties: { host: { type: 'string', required: true } },
    }, { host: 'smtp.local' })).toEqual([])
  })

  it('rejects type and enum mismatches', () => {
    expect(validateSettingsData({
      type: 'object',
      properties: {
        port: { type: 'integer' },
        tls: { type: 'string', enum: ['none', 'tls'] },
      },
    }, { port: '2525', tls: 'starttls' })).toEqual(['port', 'tls'])
  })
})

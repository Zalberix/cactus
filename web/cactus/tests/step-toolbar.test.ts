import { describe, expect, it } from 'vitest'
import {
  filterStepCatalog,
  schemaChoiceOptions,
  selectedSchemaIdForStep,
  shouldPromptForSchemaChoice,
} from '../app/components/dag/step-toolbar-utils'

describe('step toolbar catalog', () => {
  it('searches tasks and control steps together', () => {
    const result = filterStepCatalog([
      { name: 'Email SMTP', category: 'Channels', stepType: 'task', available: true },
      { name: 'IF Condition', category: 'Logic', stepType: 'control', available: true },
    ], 'if')

    expect(result.map(i => i.name)).toEqual(['IF Condition'])
  })

  it('keeps only schemas implemented by workers', () => {
    const choices = schemaChoiceOptions([
      {
        id: 1,
        version: 'unused',
        created_at: '2026-05-11T09:00:00Z',
        settings_schema: { type: 'object', properties: { unused: { type: 'string' } } },
        worker_count: 0,
        ready_workers: 0,
      },
      {
        id: 2,
        version: 'active',
        created_at: '2026-05-11T10:00:00Z',
        settings_schema: { type: 'object', properties: { host: { type: 'string' } } },
        worker_count: 1,
        ready_workers: 1,
      },
    ])

    expect(choices.map(schema => schema.id)).toEqual([2])
  })

  it('prompts only when a task has multiple worker-backed schemas and no selected schema', () => {
    const schemas = [
      {
        id: 1,
        version: 'one',
        created_at: '2026-05-11T09:00:00Z',
        settings_schema: { type: 'object', properties: {} },
        worker_count: 1,
        ready_workers: 1,
      },
      {
        id: 2,
        version: 'two',
        created_at: '2026-05-11T10:00:00Z',
        settings_schema: { type: 'object', properties: {} },
        worker_count: 1,
        ready_workers: 1,
      },
    ]

    expect(shouldPromptForSchemaChoice('task', schemas)).toBe(true)
    expect(shouldPromptForSchemaChoice('task', schemas, 2)).toBe(false)
    expect(shouldPromptForSchemaChoice('control', schemas)).toBe(false)
    expect(selectedSchemaIdForStep('task', [schemas[0]])).toBe(1)
  })
})

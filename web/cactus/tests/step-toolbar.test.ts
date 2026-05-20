import { describe, expect, it } from 'vitest'
import {
  filterStepCatalog,
  schemaChoiceOptions,
  schemaChoiceTitle,
  schemaWorkerName,
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

  it('uses the service name for schema choice titles', () => {
    expect(schemaChoiceTitle({ version: 'v2', worker_name: 'smtp-worker' }, 'SMTP Email')).toBe('smtp-worker v2')
    expect(schemaChoiceTitle({ version: 'v2' }, 'SMTP Email')).toBe('SMTP Email v2')
    expect(schemaChoiceTitle({ version: 'v2', worker_name: '  ' })).toBe('Schema v2')
  })

  it('finds the selected schema worker name', () => {
    const schemas = [
      {
        id: 1,
        version: 'one',
        created_at: '2026-05-11T09:00:00Z',
        settings_schema: { type: 'object', properties: {} },
        worker_name: 'smtp-worker',
        worker_count: 1,
        ready_workers: 1,
      },
    ]

    expect(schemaWorkerName(schemas, 1)).toBe('smtp-worker')
    expect(schemaWorkerName(schemas, 2)).toBeUndefined()
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

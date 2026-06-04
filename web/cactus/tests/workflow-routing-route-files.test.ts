import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('workflow routing page files', () => {
  it('keeps routing child pages from rendering through routing.vue as a nested parent', () => {
    const pagesRoot = resolve(__dirname, '../app/pages/org/[orgId]/workflows/[workflowId]')
    const inputSchemasIndex = resolve(pagesRoot, 'routing/input-schemas/index.vue')
    const inputSchemaIndex = resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/index.vue')

    expect(existsSync(resolve(pagesRoot, 'routing.vue'))).toBe(false)
    expect(existsSync(resolve(pagesRoot, 'routing/index.vue'))).toBe(true)
    expect(existsSync(inputSchemasIndex)).toBe(true)
    expect(readFileSync(inputSchemasIndex, 'utf8')).toContain('navigateTo')
    expect(existsSync(inputSchemaIndex)).toBe(true)
    expect(readFileSync(inputSchemaIndex, 'utf8')).toContain('navigateTo')
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/new.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/edit.vue'))).toBe(true)
  })
})

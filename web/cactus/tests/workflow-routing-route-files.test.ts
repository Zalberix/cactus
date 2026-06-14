import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('workflow routing page files', () => {
  it('keeps routing child pages from rendering through routing.vue as a nested parent', () => {
    const pagesRoot = resolve(__dirname, '../app/pages/org/[orgId]/workflows/[workflowId]')

    expect(existsSync(resolve(pagesRoot, 'routing.vue'))).toBe(false)
    expect(existsSync(resolve(pagesRoot, 'routing/index.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/testing/index.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/new.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/edit.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/compatibilities/index.vue'))).toBe(true)
    expect(existsSync(resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/compatibilities/new.vue'))).toBe(true)

    const inputSchemasIndex = resolve(pagesRoot, 'routing/input-schemas/index.vue')
    const inputSchemaIndex = resolve(pagesRoot, 'routing/input-schemas/[inputSchemaId]/index.vue')
    expect(existsSync(inputSchemasIndex)).toBe(true)
    expect(readFileSync(inputSchemasIndex, 'utf8')).toContain('navigateTo')
    expect(existsSync(inputSchemaIndex)).toBe(true)
    expect(readFileSync(inputSchemaIndex, 'utf8')).toContain('navigateTo')
  })

  it('keeps the testing title in the experiments list header only', () => {
    const componentRoot = resolve(__dirname, '../app/components/dag/routing')
    const testingPage = readFileSync(resolve(componentRoot, 'RoutingTestingPage.vue'), 'utf8')
    const experimentsList = readFileSync(resolve(componentRoot, 'RoutingExperimentsList.vue'), 'utf8')

    expect(testingPage).not.toContain("t('workflowRouting.testing')")
    expect(experimentsList).toContain("t('workflowRouting.testing')")
    expect(experimentsList).toContain("t('workflowRouting.createTesting')")
    expect(experimentsList).toContain('<Card class="border-0 shadow-none">')
  })

  it('renders a plain routing header with create schema before testing', () => {
    const pagesRoot = resolve(__dirname, '../app/pages/org/[orgId]/workflows/[workflowId]')
    const routingIndex = readFileSync(resolve(pagesRoot, 'routing/index.vue'), 'utf8')

    expect(routingIndex).not.toContain('Route class="h-6 w-6"')
    expect(routingIndex).not.toContain('rounded-3xl border')
    expect(routingIndex.indexOf("t('workflowRouting.createInputSchema')")).toBeLessThan(
      routingIndex.indexOf("t('workflowRouting.testing')"),
    )
  })
})

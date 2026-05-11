import { describe, expect, it } from 'vitest'
import { workflowOverviewPath } from '../app/composables/useWorkflows'

describe('workflow version overview routing', () => {
  it('links workflow rows to the version overview page', () => {
    expect(workflowOverviewPath(12, 34)).toBe('/org/12/workflows/34')
  })
})

import { describe, expect, it } from 'vitest'
import { workflowOverviewPath, workflowVersionEditorPath } from '../app/composables/useWorkflows'

describe('workflow version overview routing', () => {
  it('links workflow rows to the version overview page', () => {
    expect(workflowOverviewPath(12, 34)).toBe('/org/12/workflows/34')
  })

  it('links workflow versions to the version editor route', () => {
    expect(workflowVersionEditorPath(12, 34, 56)).toBe('/org/12/workflows/34/versions/56/edit')
  })
})

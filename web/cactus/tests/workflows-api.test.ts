import { describe, expect, it } from 'vitest'
import { workflowVersionCount } from '../app/composables/useWorkflows'

describe('workflow API composable', () => {
  it('uses total workflow version count before the legacy active count', () => {
    expect(workflowVersionCount({ version_count: 3, active_version_count: 0 })).toBe(3)
    expect(workflowVersionCount({ active_version_count: 2 })).toBe(2)
    expect(workflowVersionCount({})).toBe(0)
  })
})

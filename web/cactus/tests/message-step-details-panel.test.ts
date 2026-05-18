import { mount } from '@vue/test-utils'
import { computed } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MessageStepDetailsPanel from '../app/components/messages/MessageStepDetailsPanel.vue'

describe('MessageStepDetailsPanel', () => {
  beforeEach(() => {
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('useI18n', () => ({
      t: (key: string) => key,
    }))
  })

  it('renders runtime duration without unresolved component warnings', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})

    mount(MessageStepDetailsPanel, {
      props: {
        step: {
          id: 4,
          step_type: 'control',
          control_kind: 'start',
          input_mapping: [],
        },
        runtime: {
          id: 6,
          step_id: 4,
          status: 'completed',
          outcome: 'success',
          started_at: '2026-05-18T22:08:55.295501Z',
          completed_at: '2026-05-18T22:08:56.295501Z',
        },
        messageValue: {},
        runSteps: new Map(),
      },
    })

    expect(warn.mock.calls.flat().join('\n')).not.toContain('Failed to resolve component')
    warn.mockRestore()
  })
})

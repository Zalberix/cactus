import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import RunDuration from '../app/components/messages/RunDuration.vue'

describe('RunDuration', () => {
  it('renders server-provided duration', () => {
    const wrapper = mount(RunDuration, {
      props: {
        durationMs: 1500,
      },
    })

    expect(wrapper.text()).toBe('1s')
  })

  it('does not derive live duration on the client', () => {
    const wrapper = mount(RunDuration, {
      props: {
        startedAt: '2026-05-18T08:00:00Z',
      },
    })

    expect(wrapper.text()).toBe('-')
  })
})

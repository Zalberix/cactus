import { mount } from '@vue/test-utils'
import { computed } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AppBreadcrumbs from '../app/components/layout/AppBreadcrumbs.vue'

const route = vi.hoisted(() => ({
  path: '/org/1/workflows/3/versions/7/edit',
}))

vi.stubGlobal('useI18n', () => ({
  t: (key: string) => key,
}))

vi.stubGlobal('useRoute', () => route)

vi.stubGlobal('useOrgStore', () => ({
  currentOrg: { name: 'Acme' },
  organizations: [{ id: 1, name: 'Acme' }],
}))

vi.stubGlobal('computed', computed)

describe('AppBreadcrumbs', () => {
  beforeEach(() => {
    route.path = '/org/1/workflows/3/versions/7/edit'
  })

  it('does not link to the unresolved workflow versions segment', () => {
    const wrapper = mount(AppBreadcrumbs, {
      global: {
        stubs: {
          NuxtLink: {
            props: ['to'],
            template: '<a :href="to"><slot /></a>',
          },
        },
      },
    })

    const links = wrapper.findAll('a').map(link => link.attributes('href'))

    expect(links).not.toContain('/org/1/workflows/3/versions')
  })
})

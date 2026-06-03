import { flushPromises, mount } from '@vue/test-utils'
import { computed, nextTick } from 'vue'
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
  currentOrg: { name: 'Test Organization' },
  organizations: [{ id: 1, name: 'Test Organization' }],
}))

const fetchWorkflow = vi.fn()
const fetchVersionSummaries = vi.fn()

vi.stubGlobal('useWorkflows', () => ({
  fetchWorkflow,
}))

vi.stubGlobal('useVersions', () => ({
  fetchVersionSummaries,
}))

vi.stubGlobal('computed', computed)

describe('AppBreadcrumbs', () => {
  beforeEach(() => {
    route.path = '/org/1/workflows/3/versions/7/edit'
    fetchWorkflow.mockResolvedValue({
      id: 3,
      name: 'Customer onboarding',
      priority: 1,
      system_id: 1,
      created_at: '',
      updated_at: '',
    })
    fetchVersionSummaries.mockResolvedValue([
      {
        id: 7,
        workflow_id: 3,
        name: 'Draft with email checks',
        version_number: 7,
        is_valid: false,
        is_active: false,
        traffic_weight: 0,
        is_control_group: false,
        run_count: 0,
        created_at: '',
      },
    ])
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

  it('shows workflow and version names as clickable breadcrumbs', async () => {
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

    await flushPromises()

    const links = wrapper.findAll('a').map(link => ({
      href: link.attributes('href'),
      text: link.text(),
    }))

    expect(links).toContainEqual({
      href: '/org/1/workflows/3',
      text: 'Customer onboarding',
    })
    expect(links).toContainEqual({
      href: '/org/1/workflows/3/versions/7/edit',
      text: 'Draft with email checks',
    })
    expect(links.some(link => link.text === '3' || link.text === '7')).toBe(false)
  })

  it('updates the cached version name when the editor reports a successful rename', async () => {
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

    await flushPromises()

    window.dispatchEvent(new CustomEvent('cactus:workflow-version-name-updated', {
      detail: {
        versionId: 7,
        name: 'Renamed release',
      },
    }))
    await nextTick()

    const versionLink = wrapper.findAll('a').find(link =>
      link.attributes('href') === '/org/1/workflows/3/versions/7/edit',
    )

    expect(versionLink?.text()).toBe('Renamed release')
  })
})

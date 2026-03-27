<script setup lang="ts">
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '~/components/ui/breadcrumb'

interface BreadcrumbEntry {
  label: string
  to?: string
}

const { t } = useI18n()
const route = useRoute()
const orgStore = useOrgStore()

const sectionLabels: Record<string, string> = {
  workflows: 'nav.workflows',
  workers: 'nav.workers',
  messages: 'nav.messages',
  settings: 'nav.settings',
  users: 'nav.users',
  roles: 'nav.roles',
  systems: 'nav.systems',
  tokens: 'nav.tokens',
}

const crumbs = computed<BreadcrumbEntry[]>(() => {
  const entries: BreadcrumbEntry[] = []
  const path = route.path

  // Extract orgId from path: /org/:orgId/...
  const orgMatch = path.match(/^\/org\/(\d+)/)
  if (!orgMatch) return entries

  const orgId = orgMatch[1]
  const orgName = orgStore.currentOrg?.name ?? `Org ${orgId}`

  entries.push({ label: orgName, to: `/org/${orgId}/workflows` })

  // Segments after /org/:orgId/
  const rest = path.replace(`/org/${orgId}`, '').split('/').filter(Boolean)

  for (let i = 0; i < rest.length; i++) {
    const seg = rest[i]
    const i18nKey = sectionLabels[seg]
    const label = i18nKey ? t(i18nKey) : seg

    if (i < rest.length - 1) {
      const to = `/org/${orgId}/${rest.slice(0, i + 1).join('/')}`
      entries.push({ label, to })
    } else {
      entries.push({ label })
    }
  }

  return entries
})
</script>

<template>
  <Breadcrumb>
    <BreadcrumbList>
      <template v-for="(crumb, idx) in crumbs" :key="idx">
        <BreadcrumbSeparator v-if="idx > 0" />
        <BreadcrumbItem>
          <BreadcrumbLink v-if="crumb.to" :href="crumb.to" as-child>
            <NuxtLink :to="crumb.to">
              {{ crumb.label }}
            </NuxtLink>
          </BreadcrumbLink>
          <BreadcrumbPage v-else class="text-primary font-medium">
            {{ crumb.label }}
          </BreadcrumbPage>
        </BreadcrumbItem>
      </template>
    </BreadcrumbList>
  </Breadcrumb>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '~/components/ui/breadcrumb'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'

interface BreadcrumbEntry {
  label: string
  to?: string
}

const { t } = useI18n()
const route = useRoute()
const orgStore = useOrgStore()
const { fetchWorkflow } = useWorkflows()
const { fetchVersionSummaries } = useVersions()

const workflowNamesById = ref<Record<number, string>>({})
const versionNamesById = ref<Record<number, string>>({})
const loadingWorkflowIds = new Set<number>()
const loadingVersionWorkflowIds = new Set<number>()

const sectionLabels: Record<string, string> = {
  workflows: 'nav.workflows',
  versions: 'editor.versions',
  workers: 'nav.workers',
  messages: 'nav.messages',
  settings: 'nav.settings',
  users: 'nav.users',
  roles: 'nav.roles',
  systems: 'nav.systems',
  tokens: 'nav.tokens',
}

function isBreadcrumbRouteLinkable(path: string) {
  return !/^\/org\/\d+\/workflows\/\d+\/versions(?:\/\d+)?$/.test(path)
}

function parseRouteIds(path: string) {
  const match = path.match(/^\/org\/(\d+)(?:\/workflows\/(\d+)(?:\/versions\/(\d+))?)?/)
  return {
    workflowId: match?.[2] ? Number(match[2]) : undefined,
    versionId: match?.[3] ? Number(match[3]) : undefined,
  }
}

function setWorkflowName(id: number, name: string) {
  workflowNamesById.value = {
    ...workflowNamesById.value,
    [id]: name,
  }
}

function setVersionNames(versions: Array<{ id: number, name?: string, version_number: number }>) {
  versionNamesById.value = {
    ...versionNamesById.value,
    ...Object.fromEntries(
      versions.map(version => [
        version.id,
        version.name || t('editor.versionNumber', { number: version.version_number }),
      ]),
    ),
  }
}

async function loadRouteNames(path: string) {
  const { workflowId, versionId } = parseRouteIds(path)

  if (workflowId && !workflowNamesById.value[workflowId] && !loadingWorkflowIds.has(workflowId)) {
    loadingWorkflowIds.add(workflowId)
    try {
      const workflow = await fetchWorkflow(workflowId)
      setWorkflowName(workflowId, workflow.name)
    }
    catch {
      // Breadcrumbs should degrade to URL labels if contextual data is unavailable.
    }
    finally {
      loadingWorkflowIds.delete(workflowId)
    }
  }

  if (workflowId && versionId && !versionNamesById.value[versionId] && !loadingVersionWorkflowIds.has(workflowId)) {
    loadingVersionWorkflowIds.add(workflowId)
    try {
      const versions = await fetchVersionSummaries(workflowId)
      setVersionNames(versions)
    }
    catch {
      // Breadcrumbs should degrade to URL labels if contextual data is unavailable.
    }
    finally {
      loadingVersionWorkflowIds.delete(workflowId)
    }
  }
}

watch(
  () => route.path,
  path => loadRouteNames(path),
  { immediate: true },
)

const crumbs = computed<BreadcrumbEntry[]>(() => {
  const entries: BreadcrumbEntry[] = []
  const path = route.path

  // Extract orgId from path: /org/:orgId/...
  const orgMatch = path.match(/^\/org\/(\d+)/)
  if (!orgMatch) return entries

  const orgId = orgMatch[1]
  const orgName = orgStore.currentOrg?.name ?? (orgStore.organizations.length > 0 ? `Org ${orgId}` : '...')

  entries.push({ label: orgName, to: `/org/${orgId}/workflows` })

  // Segments after /org/:orgId/
  const rest = path.replace(`/org/${orgId}`, '').split('/').filter(Boolean)

  for (let i = 0; i < rest.length; i++) {
    const seg = rest[i]
    const i18nKey = sectionLabels[seg]
    const previous = rest[i - 1]
    const next = rest[i + 1]
    const isWorkflowId = previous === 'workflows' && /^\d+$/.test(seg)
    const isVersionId = previous === 'versions' && /^\d+$/.test(seg)
    const label = isWorkflowId
      ? workflowNamesById.value[Number(seg)] ?? seg
      : isVersionId
        ? versionNamesById.value[Number(seg)] ?? t('editor.versionNumber', { number: seg })
        : i18nKey ? t(i18nKey) : seg

    if (i < rest.length - 1) {
      const rawTo = `/org/${orgId}/${rest.slice(0, i + 1).join('/')}`
      const to = isVersionId && next === 'edit'
        ? workflowVersionEditorPath(Number(orgId), Number(rest[i - 2]), Number(seg))
        : rawTo
      entries.push({ label, to: isBreadcrumbRouteLinkable(to) ? to : undefined })
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
          <BreadcrumbLink v-if="crumb.to" as-child>
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

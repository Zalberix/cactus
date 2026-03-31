import type { ApiResponse } from '~/utils/api-types'

export const useOrgStore = defineStore('org', () => {
  const currentOrgId = ref<number | null>(null)
  const organizations = ref<Organization[]>([])

  const currentOrg = computed(() =>
    organizations.value.find(o => o.id === currentOrgId.value),
  )

  // Restore persisted org selection on init
  if (import.meta.client) {
    const saved = localStorage.getItem('cactus_current_org_id')
    if (saved) {
      currentOrgId.value = Number(saved)
    }
  }

  async function fetchOrganizations() {
    const { api } = useApi()
    const resp = await api<ApiResponse<Organization[]>>('/organizations')

    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch organizations')
    }

    organizations.value = resp.data

    // If no org selected yet, pick the first one
    if (currentOrgId.value === null && resp.data.length > 0) {
      switchOrg(resp.data[0].id)
    }
  }

  function switchOrg(orgId: number) {
    currentOrgId.value = orgId
    if (import.meta.client) {
      localStorage.setItem('cactus_current_org_id', String(orgId))
    }
  }

  return {
    currentOrgId,
    organizations,
    currentOrg,
    fetchOrganizations,
    switchOrg,
  }
})

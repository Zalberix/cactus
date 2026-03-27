import type { ApiResponse, PaginationMeta } from '~/utils/api-types'

export interface Organization {
  id: number
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export function useOrganizations() {
  const { api } = useApi()

  async function fetchOrgs(page: number, perPage = 20) {
    const resp = await api<ApiResponse<Organization[]>>(
      `/organizations?page=${page}&per_page=${perPage}`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch organizations')
    }
    return { data: resp.data, meta: resp.meta as PaginationMeta }
  }

  async function createOrg(name: string, description?: string) {
    const resp = await api<ApiResponse<Organization>>('/organizations', {
      method: 'POST',
      body: { name, description },
    })
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create organization')
    }
    return resp.data
  }

  async function updateOrg(orgId: number, name: string, description?: string) {
    const resp = await api<ApiResponse<Organization>>(`/organizations/${orgId}`, {
      method: 'PUT',
      body: { name, description },
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update organization')
    }
    return resp.data
  }

  async function deleteOrg(orgId: number) {
    const resp = await api<ApiResponse<null>>(`/organizations/${orgId}`, {
      method: 'DELETE',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete organization')
    }
  }

  return { fetchOrgs, createOrg, updateOrg, deleteOrg }
}

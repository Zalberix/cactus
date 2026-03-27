import type { ApiResponse } from '~/utils/api-types'

export interface Role {
  id: number
  name: string
  description?: string
  permissions: string[]
}

export interface CreateRoleRequest {
  name: string
  description?: string
  permissions: string[]
}

export function useRoles() {
  const { api } = useApi()

  async function fetchRoles(orgId: number) {
    const resp = await api<ApiResponse<Role[]>>(
      `/organizations/${orgId}/roles`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch roles')
    }
    return resp.data
  }

  async function createRole(orgId: number, data: CreateRoleRequest) {
    const resp = await api<ApiResponse<Role>>(
      `/organizations/${orgId}/roles`,
      {
        method: 'POST',
        body: data,
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create role')
    }
    return resp.data
  }

  async function updateRole(roleId: number, data: Partial<CreateRoleRequest>) {
    const resp = await api<ApiResponse<Role>>(`/roles/${roleId}`, {
      method: 'PUT',
      body: data,
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update role')
    }
    return resp.data
  }

  async function deleteRole(roleId: number) {
    const resp = await api<ApiResponse<null>>(`/roles/${roleId}`, {
      method: 'DELETE',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete role')
    }
  }

  async function fetchRolePermissions(roleId: number) {
    const resp = await api<ApiResponse<string[]>>(
      `/roles/${roleId}/permissions`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch role permissions')
    }
    return resp.data
  }

  async function assignRole(roleId: number, userId: number) {
    const resp = await api<ApiResponse<null>>(
      `/roles/${roleId}/users/${userId}`,
      { method: 'POST' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to assign role')
    }
  }

  async function removeRole(roleId: number, userId: number) {
    const resp = await api<ApiResponse<null>>(
      `/roles/${roleId}/users/${userId}`,
      { method: 'DELETE' },
    )
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to remove role')
    }
  }

  return {
    fetchRoles,
    createRole,
    updateRole,
    deleteRole,
    fetchRolePermissions,
    assignRole,
    removeRole,
  }
}

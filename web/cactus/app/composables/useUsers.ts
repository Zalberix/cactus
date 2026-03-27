import type { ApiResponse, PaginationMeta } from '~/utils/api-types'

export interface User {
  id: number
  email: string
  name: string
  created_at: string
  roles?: { id: number; name: string }[]
}

export interface CreateUserRequest {
  email: string
  name: string
  password: string
  role_id?: number
}

export function useUsers() {
  const { api } = useApi()

  async function fetchUsers(orgId: number, page: number, perPage = 20) {
    const resp = await api<ApiResponse<User[]>>(
      `/organizations/${orgId}/users?page=${page}&per_page=${perPage}`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch users')
    }
    return { data: resp.data, meta: resp.meta as PaginationMeta }
  }

  async function createUser(orgId: number, data: CreateUserRequest) {
    const resp = await api<ApiResponse<User>>(
      `/organizations/${orgId}/users`,
      {
        method: 'POST',
        body: data,
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create user')
    }
    return resp.data
  }

  async function updateUser(userId: number, data: Partial<{ name: string; email: string }>) {
    const resp = await api<ApiResponse<User>>(`/users/${userId}`, {
      method: 'PUT',
      body: data,
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update user')
    }
    return resp.data
  }

  async function deleteUser(userId: number) {
    const resp = await api<ApiResponse<null>>(`/users/${userId}`, {
      method: 'DELETE',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete user')
    }
  }

  return { fetchUsers, createUser, updateUser, deleteUser }
}

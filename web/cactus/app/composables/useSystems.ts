import type { ApiResponse } from '~/utils/api-types'

export interface System {
  id: number
  name: string
  description?: string
  created_at: string
  active_tokens_count?: number
}

export interface Token {
  id: number
  public_token: string
  private_token?: string
  created_at: string
  is_active: boolean
  name?: string
}

export interface CreateTokenResponse {
  id: number
  public_token: string
  private_token: string
}

export function useSystems() {
  const { api } = useApi()

  async function fetchSystems(orgId: number) {
    const resp = await api<ApiResponse<System[]>>(
      `/organizations/${orgId}/systems`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch systems')
    }
    return resp.data
  }

  async function createSystem(orgId: number, name: string, description?: string) {
    const resp = await api<ApiResponse<System>>(
      `/organizations/${orgId}/systems`,
      {
        method: 'POST',
        body: { name, description },
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create system')
    }
    return resp.data
  }

  async function updateSystem(systemId: number, name: string, description?: string) {
    const resp = await api<ApiResponse<System>>(`/systems/${systemId}`, {
      method: 'PUT',
      body: { name, description },
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to update system')
    }
    return resp.data
  }

  async function deleteSystem(systemId: number) {
    const resp = await api<ApiResponse<null>>(`/systems/${systemId}`, {
      method: 'DELETE',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to delete system')
    }
  }

  async function fetchTokens(systemId: number) {
    const resp = await api<ApiResponse<Token[]>>(
      `/systems/${systemId}/tokens`,
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to fetch tokens')
    }
    return resp.data
  }

  async function createToken(systemId: number, name?: string) {
    const resp = await api<ApiResponse<CreateTokenResponse>>(
      `/systems/${systemId}/tokens`,
      {
        method: 'POST',
        body: { name },
      },
    )
    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Failed to create token')
    }
    return resp.data
  }

  async function revokeToken(tokenId: number) {
    const resp = await api<ApiResponse<null>>(`/tokens/${tokenId}`, {
      method: 'DELETE',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to revoke token')
    }
  }

  async function activateToken(tokenId: number) {
    const resp = await api<ApiResponse<null>>(`/tokens/${tokenId}/activate`, {
      method: 'POST',
    })
    if (!resp.success) {
      throw new Error(resp.error?.message ?? 'Failed to activate token')
    }
  }

  return {
    fetchSystems,
    createSystem,
    updateSystem,
    deleteSystem,
    fetchTokens,
    createToken,
    revokeToken,
    activateToken,
  }
}

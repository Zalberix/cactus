import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useApi } from '../app/composables/useApi'

type FetchOptions = {
  headers?: HeadersInit
  [key: string]: unknown
}

type ResponseErrorContext = {
  request: string
  options: FetchOptions
  response: {
    status: number
    _data?: unknown
  }
}

type FetchHooks = {
  onRequest?: (ctx: { options: FetchOptions }) => void
  onResponseError?: (ctx: ResponseErrorContext) => void | Promise<void>
}

function unauthorizedError() {
  const response = {
    status: 401,
    _data: {
      success: false,
      error: { message: 'Invalid token' },
    },
  }
  const error = new Error('Invalid token') as Error & {
    response: typeof response
    status: number
    statusCode: number
  }
  error.response = response
  error.status = 401
  error.statusCode = 401
  return error
}

describe('useApi auth refresh', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('refreshes the access token and retries the original request after an invalid token response', async () => {
    const authStore = {
      accessToken: 'expired-token' as string | null,
      get isAuthenticated() {
        return !!this.accessToken
      },
      silentRefresh: vi.fn(async () => {
        authStore.accessToken = 'fresh-token'
      }),
      logout: vi.fn(),
    }
    const calls: Array<{ request: string, authorization: string | null }> = []

    vi.stubGlobal('useAuthStore', () => authStore)
    vi.stubGlobal('useRuntimeConfig', () => ({ public: { apiBase: '/api/v1' } }))
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))

    const fetch = vi.fn()
    fetch.create = vi.fn((hooks: FetchHooks) => async (request: string, options: FetchOptions = {}) => {
      const requestOptions = { ...options }
      hooks.onRequest?.({ options: requestOptions })

      calls.push({
        request,
        authorization: new Headers(requestOptions.headers).get('Authorization'),
      })

      if (calls.length === 1) {
        const error = unauthorizedError()
        await hooks.onResponseError?.({
          request,
          options: requestOptions,
          response: error.response,
        })
        throw error
      }

      return { success: true, data: { ok: true } }
    })
    vi.stubGlobal('$fetch', fetch)

    const { api } = useApi()

    await expect(api('/protected')).resolves.toEqual({
      success: true,
      data: { ok: true },
    })
    expect(authStore.silentRefresh).toHaveBeenCalledTimes(1)
    expect(authStore.logout).not.toHaveBeenCalled()
    expect(calls).toEqual([
      { request: '/protected', authorization: 'Bearer expired-token' },
      { request: '/protected', authorization: 'Bearer fresh-token' },
    ])
  })

  it('logs out when the token cannot be refreshed', async () => {
    const authStore = {
      accessToken: 'expired-token' as string | null,
      get isAuthenticated() {
        return !!this.accessToken
      },
      silentRefresh: vi.fn(async () => {
        authStore.accessToken = null
      }),
      logout: vi.fn(),
    }

    vi.stubGlobal('useAuthStore', () => authStore)
    vi.stubGlobal('useRuntimeConfig', () => ({ public: { apiBase: '/api/v1' } }))
    vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))

    const fetch = vi.fn()
    fetch.create = vi.fn((hooks: FetchHooks) => async (request: string, options: FetchOptions = {}) => {
      const requestOptions = { ...options }
      hooks.onRequest?.({ options: requestOptions })
      const error = unauthorizedError()
      await hooks.onResponseError?.({
        request,
        options: requestOptions,
        response: error.response,
      })
      throw error
    })
    vi.stubGlobal('$fetch', fetch)

    const { api } = useApi()

    await expect(api('/protected')).rejects.toThrow('Invalid token')
    expect(authStore.silentRefresh).toHaveBeenCalledTimes(1)
    expect(authStore.logout).toHaveBeenCalledTimes(1)
  })
})

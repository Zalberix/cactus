import type { ApiResponse, LoginResponse } from '~/utils/api-types'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(null)
  const user = ref<{ id: number; email: string; name: string } | null>(null)

  const isAuthenticated = computed(() => !!accessToken.value)

  const refreshCookie = useCookie('refresh_token', {
    sameSite: 'strict',
    maxAge: 7 * 24 * 3600,
  })

  let refreshPromise: Promise<void> | null = null

  async function login(email: string, password: string) {
    const config = useRuntimeConfig()
    const resp = await $fetch<ApiResponse<LoginResponse>>(
      `${config.public.apiBase}/auth/login`,
      {
        method: 'POST',
        body: { email, password },
      },
    )

    if (!resp.success || !resp.data) {
      throw new Error(resp.error?.message ?? 'Login failed')
    }

    accessToken.value = resp.data.access_token
    refreshCookie.value = resp.data.refresh_token
  }

  async function silentRefresh(): Promise<void> {
    if (refreshPromise) return refreshPromise

    refreshPromise = (async () => {
      try {
        const token = refreshCookie.value
        if (!token) {
          accessToken.value = null
          return
        }

        const config = useRuntimeConfig()
        const resp = await $fetch<ApiResponse<LoginResponse>>(
          `${config.public.apiBase}/auth/refresh`,
          {
            method: 'POST',
            body: { refresh_token: token },
          },
        )

        if (resp.success && resp.data) {
          accessToken.value = resp.data.access_token
          refreshCookie.value = resp.data.refresh_token
        }
        else {
          accessToken.value = null
          refreshCookie.value = null
        }
      }
      catch {
        accessToken.value = null
        refreshCookie.value = null
      }
      finally {
        refreshPromise = null
      }
    })()

    return refreshPromise
  }

  function logout() {
    accessToken.value = null
    user.value = null
    refreshCookie.value = null
    navigateTo('/login')
  }

  return {
    accessToken,
    user,
    isAuthenticated,
    login,
    silentRefresh,
    logout,
  }
})

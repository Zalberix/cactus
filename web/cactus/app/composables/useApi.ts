import { toast } from '~/components/ui/toast/use-toast'

type FetchErrorLike = {
  response?: {
    status?: number
  }
  status?: number
  statusCode?: number
}

function isUnauthorizedError(error: unknown) {
  const fetchError = error as FetchErrorLike
  return (
    fetchError.response?.status === 401
    || fetchError.status === 401
    || fetchError.statusCode === 401
  )
}

export function useApi() {
  const authStore = useAuthStore()
  const config = useRuntimeConfig()
  const { t } = useI18n()

  const rawApi = $fetch.create({
    baseURL: config.public.apiBase,

    onRequest({ options }) {
      const token = authStore.accessToken
      if (token) {
        const headers = new Headers(options.headers)
        headers.set('Authorization', `Bearer ${token}`)
        options.headers = headers
      }
    },

    onResponseError({ response }) {
      if (response.status === 403) {
        toast({
          title: getErrorMessage(response._data, t('error.forbidden')),
          variant: 'destructive',
        })
        return
      }

      if (response.status >= 500) {
        toast({
          title: getErrorMessage(response._data, t('error.server')),
          variant: 'destructive',
        })
      }
    },
  })

  async function api<T>(
    request: Parameters<typeof rawApi>[0],
    options?: Parameters<typeof rawApi>[1],
  ): Promise<T> {
    try {
      return await rawApi<T>(request, options)
    }
    catch (error) {
      if (!isUnauthorizedError(error)) throw error

      try {
        await authStore.silentRefresh()
      }
      catch {
        authStore.logout()
        throw error
      }

      if (!authStore.isAuthenticated || !authStore.accessToken) {
        authStore.logout()
        throw error
      }

      return await rawApi<T>(request, options)
    }
  }

  return { api }
}

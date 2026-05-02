import { toast } from '~/components/ui/toast/use-toast'

export function useApi() {
  const authStore = useAuthStore()
  const config = useRuntimeConfig()
  const { t } = useI18n()

  const api = $fetch.create({
    baseURL: config.public.apiBase,

    onRequest({ options }) {
      const token = authStore.accessToken
      if (token) {
        const headers = new Headers(options.headers)
        headers.set('Authorization', `Bearer ${token}`)
        options.headers = headers
      }
    },

    async onResponseError({ response, request, options }) {
      if (response.status === 401) {
        try {
          await authStore.silentRefresh()

          if (authStore.isAuthenticated) {
            // Retry the original request with the new token
            const headers = new Headers(options.headers)
            headers.set('Authorization', `Bearer ${authStore.accessToken}`)
            options.headers = headers
            return $fetch(request, {
              ...options,
              headers,
            })
          }
        }
        catch {
          // silentRefresh already navigates to /login on failure
        }
        return
      }

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

  return { api }
}

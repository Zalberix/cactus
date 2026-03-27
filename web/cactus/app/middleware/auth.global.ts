export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()

  const publicRoutes = ['/login']

  // If navigating to a public route while authenticated, redirect to home
  if (publicRoutes.includes(to.path)) {
    if (auth.isAuthenticated) {
      return navigateTo('/')
    }
    return
  }

  // Protected routes: check authentication
  if (!auth.isAuthenticated) {
    // Attempt silent refresh before redirecting
    try {
      await auth.silentRefresh()
    }
    catch {
      // silentRefresh failed
    }

    // If still not authenticated after refresh attempt, redirect to login
    if (!auth.isAuthenticated) {
      return navigateTo('/login')
    }
  }
})

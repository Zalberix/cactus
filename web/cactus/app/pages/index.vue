<script setup lang="ts">
const orgStore = useOrgStore()

onMounted(async () => {
  try {
    await orgStore.fetchOrganizations()

    if (orgStore.currentOrgId) {
      await navigateTo(`/org/${orgStore.currentOrgId}/workflows`, { replace: true })
    }
    else if (orgStore.organizations.length > 0) {
      orgStore.switchOrg(orgStore.organizations[0].id)
      await navigateTo(`/org/${orgStore.organizations[0].id}/workflows`, { replace: true })
    }
  }
  catch {
    // If org fetch fails, stay on index page
  }
})
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div class="text-center">
      <h1 class="text-2xl font-semibold text-foreground">
        Cactus
      </h1>
      <p class="mt-2 text-sm text-muted-foreground">
        {{ $t('common.loading') }}
      </p>
    </div>
  </div>
</template>

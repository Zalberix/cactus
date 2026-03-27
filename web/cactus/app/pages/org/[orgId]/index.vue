<script setup lang="ts">
const route = useRoute()
const orgId = route.params.orgId as string

await navigateTo(`/org/${orgId}/workflows`, { replace: true })
</script>

<template>
  <div />
</template>

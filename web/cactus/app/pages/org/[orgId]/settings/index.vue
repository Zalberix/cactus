<script setup lang="ts">
const route = useRoute()

definePageMeta({
  layout: 'default',
})

await navigateTo(`/org/${route.params.orgId}/settings/users`, { replace: true })
</script>

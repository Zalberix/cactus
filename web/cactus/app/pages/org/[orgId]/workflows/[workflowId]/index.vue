<script setup lang="ts">
const route = useRoute()

definePageMeta({
  layout: 'default',
})

await navigateTo(`/org/${route.params.orgId}/workflows/${route.params.workflowId}/edit`, { replace: true })
</script>

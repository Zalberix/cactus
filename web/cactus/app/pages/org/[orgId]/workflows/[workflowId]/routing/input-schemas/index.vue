<script setup lang="ts">
import { workflowRoutingPath } from '~/composables/useWorkflowRouting'

const route = useRoute()

await navigateTo(
  workflowRoutingPath(Number(route.params.orgId), Number(route.params.workflowId)),
  { replace: true },
)
</script>

<template>
  <div />
</template>

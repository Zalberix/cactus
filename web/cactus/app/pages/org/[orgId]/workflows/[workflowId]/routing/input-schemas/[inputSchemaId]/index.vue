<script setup lang="ts">
import { workflowInputSchemaEditorPath } from '~/composables/useWorkflowRouting'

const route = useRoute()

await navigateTo(
  workflowInputSchemaEditorPath(
    Number(route.params.orgId),
    Number(route.params.workflowId),
    Number(route.params.inputSchemaId),
  ),
  { replace: true },
)
</script>

<template>
  <div />
</template>

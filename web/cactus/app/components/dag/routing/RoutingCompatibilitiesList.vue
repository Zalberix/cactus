<script setup lang="ts">
import type { WorkflowSchemaCompatibility } from '~/composables/useWorkflowRouting'
import { Plus } from 'lucide-vue-next'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'

defineProps<{
  compatibilities: WorkflowSchemaCompatibility[]
  saving?: boolean
  schemaLabel: (schemaId: number) => string
  versionLabel: (versionId: number) => string
}>()

const emit = defineEmits<{
  create: []
  edit: [compatibility: WorkflowSchemaCompatibility]
  makeDefault: [compatibility: WorkflowSchemaCompatibility]
  deactivate: [compatibility: WorkflowSchemaCompatibility]
  delete: [compatibility: WorkflowSchemaCompatibility]
}>()

const { t } = useI18n()
</script>

<template>
  <Card>
    <CardHeader class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle>{{ t('workflowRouting.compatibilities') }}</CardTitle>
          <CardDescription>{{ t('workflowRouting.compatibilitiesDescription') }}</CardDescription>
        </div>
        <Button type="button" class="gap-2" :disabled="saving" data-testid="open-create-compatibility-modal" @click="emit('create')">
          <Plus class="h-4 w-4" />
          {{ t('workflowRouting.createCompatibility') }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-3">
      <div v-for="compatibility in compatibilities" :key="compatibility.id" class="rounded-2xl border p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ schemaLabel(compatibility.workflow_input_schema_id) }} to {{ versionLabel(compatibility.workflow_version_id) }}</span>
              <Badge v-if="compatibility.is_default_route">{{ t('workflowRouting.defaultRoute') }}</Badge>
              <Badge variant="outline">{{ compatibility.compatibility_type }}</Badge>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">
              Mapper {{ compatibility.workflow_input_mapper_id || 'none' }} &middot; compatibility #{{ compatibility.id }}
            </p>
          </div>
          <div class="flex gap-2">
            <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-compatibility-${compatibility.id}`" @click="emit('edit', compatibility)">
              {{ t('common.edit') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="compatibility.is_default_route || saving" @click="emit('makeDefault', compatibility)">
              {{ t('workflowRouting.makeDefaultRoute') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="!compatibility.is_active || saving" @click="emit('deactivate', compatibility)">
              {{ t('workflowRouting.deactivate') }}
            </Button>
            <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-compatibility-${compatibility.id}`" @click="emit('delete', compatibility)">
              {{ t('common.delete') }}
            </Button>
          </div>
        </div>
        <pre class="mt-3 max-h-32 overflow-auto rounded-xl bg-muted p-3 text-xs">{{ JSON.stringify(compatibility.default_values, null, 2) }}</pre>
      </div>
      <div v-if="compatibilities.length === 0" class="rounded-2xl border border-dashed p-6 text-sm text-muted-foreground">
        {{ t('workflowRouting.noCompatibilities') }}
      </div>
    </CardContent>
  </Card>
</template>

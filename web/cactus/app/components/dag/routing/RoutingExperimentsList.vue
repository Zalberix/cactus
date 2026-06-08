<script setup lang="ts">
import type {
  WorkflowExperiment,
  WorkflowExperimentScope,
  WorkflowExperimentVariant,
} from '~/composables/useWorkflowRouting'
import { GitBranch, Plus } from 'lucide-vue-next'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'

defineProps<{
  experiments: WorkflowExperiment[]
  scopesByExperiment: Record<number, WorkflowExperimentScope[]>
  variantsByScope: Record<number, WorkflowExperimentVariant[]>
  saving?: boolean
  schemaLabel: (schemaId: number) => string
  versionLabel: (versionId: number) => string
}>()

const emit = defineEmits<{
  create: []
  createScope: [experiment: WorkflowExperiment]
  edit: [experiment: WorkflowExperiment]
  activate: [experiment: WorkflowExperiment]
  pause: [experiment: WorkflowExperiment]
  delete: [experiment: WorkflowExperiment]
  createVariant: [scope: WorkflowExperimentScope]
  editScope: [experiment: WorkflowExperiment, scope: WorkflowExperimentScope]
  deleteScope: [scope: WorkflowExperimentScope]
  editVariant: [scope: WorkflowExperimentScope, variant: WorkflowExperimentVariant]
  deleteVariant: [variant: WorkflowExperimentVariant]
}>()

const { t } = useI18n()
</script>

<template>
  <Card>
    <CardHeader class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle class="flex items-center gap-2">
            <GitBranch class="h-5 w-5" />
            {{ t('workflowRouting.testing') }}
          </CardTitle>
          <CardDescription>{{ t('workflowRouting.testingDescription') }}</CardDescription>
        </div>
        <Button type="button" class="gap-2" :disabled="saving" data-testid="open-create-experiment-modal" @click="emit('create')">
          <Plus class="h-4 w-4" />
          {{ t('workflowRouting.createTesting') }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-4">
      <div v-for="experiment in experiments" :key="experiment.id" class="rounded-2xl border p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ experiment.name }}</span>
              <Badge variant="outline">{{ experiment.experiment_type }}</Badge>
              <Badge>{{ experiment.status }}</Badge>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">Experiment #{{ experiment.id }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button size="sm" variant="outline" :disabled="saving" :data-testid="`open-create-scope-${experiment.id}`" @click="emit('createScope', experiment)">
              <Plus class="mr-1 h-3.5 w-3.5" />
              {{ t('workflowRouting.createScope') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-experiment-${experiment.id}`" @click="emit('edit', experiment)">
              {{ t('common.edit') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="experiment.status === 'active' || saving" @click="emit('activate', experiment)">
              {{ t('workflowRouting.activate') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="experiment.status !== 'active' || saving" @click="emit('pause', experiment)">
              {{ t('workflowRouting.pause') }}
            </Button>
            <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-experiment-${experiment.id}`" @click="emit('delete', experiment)">
              {{ t('common.delete') }}
            </Button>
          </div>
        </div>

        <div class="mt-4 space-y-3">
          <div v-for="scope in scopesByExperiment[experiment.id] || []" :key="scope.id" class="rounded-xl bg-muted/50 p-3">
            <div class="flex flex-wrap items-center justify-between gap-2 text-sm">
              <span>{{ schemaLabel(scope.workflow_input_schema_id) }}</span>
              <div class="flex items-center gap-2">
                <span class="text-muted-foreground">{{ scope.traffic_percent }}% traffic &middot; {{ scope.fallback_policy }}</span>
                <Button size="sm" variant="outline" :disabled="saving" :data-testid="`open-create-variant-${scope.id}`" @click="emit('createVariant', scope)">
                  <Plus class="mr-1 h-3.5 w-3.5" />
                  {{ t('workflowRouting.createVariant') }}
                </Button>
                <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-scope-${scope.id}`" @click="emit('editScope', experiment, scope)">
                  {{ t('workflowRouting.editScope') }}
                </Button>
                <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-scope-${scope.id}`" @click="emit('deleteScope', scope)">
                  {{ t('workflowRouting.deleteScope') }}
                </Button>
              </div>
            </div>
            <div class="mt-2 grid gap-2 md:grid-cols-2">
              <div v-for="variant in variantsByScope[scope.id] || []" :key="variant.id" class="rounded-lg border bg-background p-3 text-sm">
                <div class="flex items-start justify-between gap-2">
                  <div class="font-medium">{{ versionLabel(variant.workflow_version_id) }}</div>
                  <div class="flex gap-2">
                    <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-variant-${variant.id}`" @click="emit('editVariant', scope, variant)">
                      {{ t('common.edit') }}
                    </Button>
                    <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-variant-${variant.id}`" @click="emit('deleteVariant', variant)">
                      {{ t('common.delete') }}
                    </Button>
                  </div>
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ variant.traffic_weight }}% &middot; {{ variant.is_control_group ? t('workflowRouting.control') : t('workflowRouting.variant') }} &middot; {{ variant.is_active ? t('status.active') : t('status.inactive') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="experiments.length === 0" class="rounded-2xl border border-dashed p-6 text-sm text-muted-foreground">
        {{ t('workflowRouting.noTesting') }}
      </div>
    </CardContent>
  </Card>
</template>

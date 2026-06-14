<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type {
  WorkflowExperiment,
  WorkflowExperimentScope,
  WorkflowExperimentVariant,
  WorkflowInputSchemaRecord,
} from '~/composables/useWorkflowRouting'
import type { RoutingDeleteTarget } from '~/components/dag/routing/types'
import { ArrowLeft } from 'lucide-vue-next'
import { workflowRoutingPath } from '~/composables/useWorkflowRouting'
import RoutingDeleteDialog from '~/components/dag/routing/RoutingDeleteDialog.vue'
import RoutingExperimentDialog from '~/components/dag/routing/RoutingExperimentDialog.vue'
import RoutingExperimentsList from '~/components/dag/routing/RoutingExperimentsList.vue'
import RoutingScopeDialog from '~/components/dag/routing/RoutingScopeDialog.vue'
import RoutingVariantDialog from '~/components/dag/routing/RoutingVariantDialog.vue'
import { Button } from '~/components/ui/button'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const routing = useWorkflowRouting()
const { fetchVersionSummaries } = useVersions()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const versions = ref<VersionSummary[]>([])
const schemas = ref<WorkflowInputSchemaRecord[]>([])
const experiments = ref<WorkflowExperiment[]>([])
const scopesByExperiment = ref<Record<number, WorkflowExperimentScope[]>>({})
const variantsByScope = ref<Record<number, WorkflowExperimentVariant[]>>({})
const loading = ref(true)
const saving = ref(false)
const experimentDialogOpen = ref(false)
const scopeDialogOpen = ref(false)
const variantDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<RoutingDeleteTarget | null>(null)

const experimentForm = reactive({
  name: 'Canary route',
  experimentType: 'canary',
  status: 'draft',
})

const scopeForm = reactive({
  experimentId: '',
  inputSchemaId: '',
  trafficPercent: 100,
  fallbackPolicy: 'default_route',
  fallbackVersionId: '',
  trafficConditionsJson: '{}',
})

const variantForm = reactive({
  scopeId: '',
  versionId: '',
  trafficWeight: 100,
  isControlGroup: false,
  isActive: true,
})

const editingExperimentId = ref<number | null>(null)
const editingScopeId = ref<number | null>(null)
const editingVariantId = ref<number | null>(null)

const activeVersions = computed(() => versions.value.filter(version => version.is_active && version.is_valid))
const activeSchemas = computed(() => schemas.value.filter(schema => schema.status === 'active'))

function schemaLabel(schemaId: number) {
  const schema = schemas.value.find(item => item.id === schemaId)
  return schema ? `${schema.code} v${schema.version_number}` : `schema #${schemaId}`
}

function versionLabel(versionId: number) {
  const version = versions.value.find(item => item.id === versionId)
  return version ? `${version.name || 'Version'} v${version.version_number}` : `#${versionId}`
}

function parseObjectJson(raw: string, label: string): Record<string, unknown> {
  const parsed = JSON.parse(raw || '{}')
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(`${label} must be a JSON object`)
  }
  return parsed as Record<string, unknown>
}

async function loadData() {
  loading.value = true
  try {
    const [versionList, schemaList, experimentList] = await Promise.all([
      fetchVersionSummaries(workflowId.value),
      routing.fetchInputSchemas(workflowId.value),
      routing.fetchExperiments(workflowId.value),
    ])
    versions.value = versionList
    schemas.value = schemaList
    experiments.value = experimentList
    await loadExperimentChildren(experimentList)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function loadExperimentChildren(items: WorkflowExperiment[]) {
  const scopes: Record<number, WorkflowExperimentScope[]> = {}
  const variants: Record<number, WorkflowExperimentVariant[]> = {}
  await Promise.all(items.map(async (experiment) => {
    const experimentScopes = await routing.fetchExperimentScopes(experiment.id)
    scopes[experiment.id] = experimentScopes
    await Promise.all(experimentScopes.map(async (scope) => {
      variants[scope.id] = await routing.fetchExperimentVariants(scope.id)
    }))
  }))
  scopesByExperiment.value = scopes
  variantsByScope.value = variants
}

async function runAction(action: () => Promise<unknown>, successTitle: string, fallbackError: string) {
  saving.value = true
  try {
    await action()
    toast({ title: successTitle })
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, fallbackError), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function cancelEditExperiment() {
  editingExperimentId.value = null
  experimentForm.name = 'Canary route'
  experimentForm.experimentType = 'canary'
  experimentForm.status = 'draft'
}

function openCreateExperimentDialog() {
  cancelEditExperiment()
  experimentDialogOpen.value = true
}

function startEditExperiment(experiment: WorkflowExperiment) {
  editingExperimentId.value = experiment.id
  experimentForm.name = experiment.name
  experimentForm.experimentType = experiment.experiment_type
  experimentForm.status = experiment.status
  experimentDialogOpen.value = true
}

async function onCreateExperiment() {
  await runAction(async () => {
    if (editingExperimentId.value) {
      await routing.updateExperiment(editingExperimentId.value, {
        name: experimentForm.name.trim(),
        experiment_type: experimentForm.experimentType,
      })
      await routing.updateExperimentStatus(editingExperimentId.value, experimentForm.status)
    }
    else {
      await routing.createExperiment(workflowId.value, {
        name: experimentForm.name.trim(),
        experiment_type: experimentForm.experimentType,
        status: experimentForm.status,
      })
    }
    experimentDialogOpen.value = false
    cancelEditExperiment()
  }, t('workflowRouting.createdTesting'), t('workflowRouting.errorCreateExperiment'))
}

function onActivateExperiment(experiment: WorkflowExperiment) {
  void runAction(() => routing.updateExperimentStatus(experiment.id, 'active'), t('workflowRouting.activatedTesting'), 'Failed to activate testing')
}

function onPauseExperiment(experiment: WorkflowExperiment) {
  void runAction(() => routing.updateExperimentStatus(experiment.id, 'paused'), t('workflowRouting.pausedTesting'), 'Failed to pause testing')
}

function onDeleteExperiment(experiment: WorkflowExperiment) {
  deleteTarget.value = {
    message: t('workflowRouting.confirmDeleteExperiment', { name: experiment.name }),
    successTitle: t('workflowRouting.deletedExperiment'),
    fallbackError: 'Failed to delete testing',
    action: () => routing.deleteExperiment(experiment.id),
  }
  deleteDialogOpen.value = true
}

function cancelEditScope() {
  editingScopeId.value = null
  scopeForm.experimentId = ''
  scopeForm.inputSchemaId = ''
  scopeForm.trafficPercent = 100
  scopeForm.fallbackPolicy = 'default_route'
  scopeForm.fallbackVersionId = ''
  scopeForm.trafficConditionsJson = '{}'
}

function openCreateScopeDialog(experiment?: WorkflowExperiment) {
  cancelEditScope()
  if (experiment) scopeForm.experimentId = String(experiment.id)
  scopeDialogOpen.value = true
}

function startEditScope(experiment: WorkflowExperiment, scope: WorkflowExperimentScope) {
  editingScopeId.value = scope.id
  scopeForm.experimentId = String(experiment.id)
  scopeForm.inputSchemaId = String(scope.workflow_input_schema_id)
  scopeForm.trafficPercent = scope.traffic_percent
  scopeForm.fallbackPolicy = scope.fallback_policy
  scopeForm.fallbackVersionId = scope.fallback_workflow_version_id ? String(scope.fallback_workflow_version_id) : ''
  scopeForm.trafficConditionsJson = JSON.stringify(scope.traffic_conditions ?? {}, null, 2)
  scopeDialogOpen.value = true
}

async function onCreateScope() {
  await runAction(async () => {
    const payload = {
      workflow_input_schema_id: Number(scopeForm.inputSchemaId),
      traffic_percent: Number(scopeForm.trafficPercent),
      fallback_policy: scopeForm.fallbackPolicy,
      fallback_workflow_version_id: scopeForm.fallbackVersionId ? Number(scopeForm.fallbackVersionId) : undefined,
      traffic_conditions: parseObjectJson(scopeForm.trafficConditionsJson, 'Traffic conditions'),
    }
    if (editingScopeId.value) {
      await routing.updateExperimentScope(editingScopeId.value, payload)
    }
    else {
      await routing.createExperimentScope(Number(scopeForm.experimentId), payload)
    }
    scopeDialogOpen.value = false
    cancelEditScope()
  }, t('workflowRouting.createdScope'), t('workflowRouting.errorCreateScope'))
}

function onDeleteScope(scope: WorkflowExperimentScope) {
  deleteTarget.value = {
    message: t('workflowRouting.confirmDeleteScope', { id: scope.id }),
    successTitle: t('workflowRouting.deletedScope'),
    fallbackError: 'Failed to delete scope',
    action: () => routing.deleteExperimentScope(scope.id),
  }
  deleteDialogOpen.value = true
}

function cancelEditVariant() {
  editingVariantId.value = null
  variantForm.scopeId = ''
  variantForm.versionId = ''
  variantForm.trafficWeight = 100
  variantForm.isControlGroup = false
  variantForm.isActive = true
}

function openCreateVariantDialog(scope?: WorkflowExperimentScope) {
  cancelEditVariant()
  if (scope) variantForm.scopeId = String(scope.id)
  variantDialogOpen.value = true
}

function startEditVariant(scope: WorkflowExperimentScope, variant: WorkflowExperimentVariant) {
  editingVariantId.value = variant.id
  variantForm.scopeId = String(scope.id)
  variantForm.versionId = String(variant.workflow_version_id)
  variantForm.trafficWeight = variant.traffic_weight
  variantForm.isControlGroup = variant.is_control_group
  variantForm.isActive = variant.is_active
  variantDialogOpen.value = true
}

async function onCreateVariant() {
  await runAction(async () => {
    const payload = {
      workflow_version_id: Number(variantForm.versionId),
      traffic_weight: Number(variantForm.trafficWeight),
      is_control_group: variantForm.isControlGroup,
      is_active: variantForm.isActive,
    }
    if (editingVariantId.value) {
      await routing.updateExperimentVariant(editingVariantId.value, payload)
    }
    else {
      await routing.createExperimentVariant(Number(variantForm.scopeId), payload)
    }
    variantDialogOpen.value = false
    cancelEditVariant()
  }, t('workflowRouting.createdVariant'), t('workflowRouting.errorCreateVariant'))
}

function onDeleteVariant(variant: WorkflowExperimentVariant) {
  deleteTarget.value = {
    message: t('workflowRouting.confirmDeleteVariant', { id: variant.id }),
    successTitle: t('workflowRouting.deletedVariant'),
    fallbackError: 'Failed to delete variant',
    action: () => routing.deleteExperimentVariant(variant.id),
  }
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  await runAction(deleteTarget.value.action, deleteTarget.value.successTitle, deleteTarget.value.fallbackError)
  deleteDialogOpen.value = false
  deleteTarget.value = null
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <Button variant="ghost" class="-ml-3 gap-2" @click="router.push(workflowRoutingPath(orgId, workflowId))">
      <ArrowLeft class="h-4 w-4" />
      {{ t('common.back') }}
    </Button>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>
    <RoutingExperimentsList
      v-else
      :experiments="experiments"
      :scopes-by-experiment="scopesByExperiment"
      :variants-by-scope="variantsByScope"
      :saving="saving"
      :schema-label="schemaLabel"
      :version-label="versionLabel"
      @create="openCreateExperimentDialog"
      @create-scope="openCreateScopeDialog"
      @edit="startEditExperiment"
      @activate="onActivateExperiment"
      @pause="onPauseExperiment"
      @delete="onDeleteExperiment"
      @create-variant="openCreateVariantDialog"
      @edit-scope="startEditScope"
      @delete-scope="onDeleteScope"
      @edit-variant="startEditVariant"
      @delete-variant="onDeleteVariant"
    />

    <RoutingExperimentDialog
      v-model:open="experimentDialogOpen"
      :form="experimentForm"
      :saving="saving"
      :editing-id="editingExperimentId"
      @submit="onCreateExperiment"
      @cancel="cancelEditExperiment"
    />

    <RoutingScopeDialog
      v-model:open="scopeDialogOpen"
      :form="scopeForm"
      :experiments="experiments"
      :active-schemas="activeSchemas"
      :saving="saving"
      :editing-id="editingScopeId"
      :schema-label="schemaLabel"
      @submit="onCreateScope"
      @cancel="cancelEditScope"
    />

    <RoutingVariantDialog
      v-model:open="variantDialogOpen"
      :form="variantForm"
      :experiments="experiments"
      :scopes-by-experiment="scopesByExperiment"
      :active-versions="activeVersions"
      :saving="saving"
      :editing-id="editingVariantId"
      :version-label="versionLabel"
      :schema-label="schemaLabel"
      @submit="onCreateVariant"
      @cancel="cancelEditVariant"
    />

    <RoutingDeleteDialog
      v-model:open="deleteDialogOpen"
      :target="deleteTarget"
      :saving="saving"
      @cancel="deleteDialogOpen = false"
      @confirm="confirmDelete"
    />
  </div>
</template>

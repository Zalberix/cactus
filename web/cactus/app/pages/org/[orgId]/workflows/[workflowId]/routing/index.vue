<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type {
  WorkflowExperiment,
  WorkflowExperimentScope,
  WorkflowExperimentVariant,
  WorkflowInputMapperRecord,
  WorkflowInputSchemaRecord,
  WorkflowSchemaCompatibility,
} from '~/composables/useWorkflowRouting'
import type { RoutingDeleteTarget } from '~/components/dag/routing/types'
import { ArrowLeft, Route } from 'lucide-vue-next'
import { workflowInputSchemaCreatePath, workflowInputSchemaEditorPath } from '~/composables/useWorkflowRouting'
import RoutingCompatibilitiesList from '~/components/dag/routing/RoutingCompatibilitiesList.vue'
import RoutingCompatibilityDialog from '~/components/dag/routing/RoutingCompatibilityDialog.vue'
import RoutingDeleteDialog from '~/components/dag/routing/RoutingDeleteDialog.vue'
import RoutingExperimentDialog from '~/components/dag/routing/RoutingExperimentDialog.vue'
import RoutingExperimentsList from '~/components/dag/routing/RoutingExperimentsList.vue'
import RoutingInputMapperDialog from '~/components/dag/routing/RoutingInputMapperDialog.vue'
import RoutingInputMappersList from '~/components/dag/routing/RoutingInputMappersList.vue'
import RoutingInputSchemasList from '~/components/dag/routing/RoutingInputSchemasList.vue'
import RoutingScopeDialog from '~/components/dag/routing/RoutingScopeDialog.vue'
import RoutingVariantDialog from '~/components/dag/routing/RoutingVariantDialog.vue'
import { Button } from '~/components/ui/button'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const { fetchWorkflow } = useWorkflows()
const { fetchVersionSummaries } = useVersions()
const routing = useWorkflowRouting()

const workflowName = ref('')
const versions = ref<VersionSummary[]>([])
const schemas = ref<WorkflowInputSchemaRecord[]>([])
const mappers = ref<WorkflowInputMapperRecord[]>([])
const compatibilities = ref<WorkflowSchemaCompatibility[]>([])
const experiments = ref<WorkflowExperiment[]>([])
const scopesByExperiment = ref<Record<number, WorkflowExperimentScope[]>>({})
const variantsByScope = ref<Record<number, WorkflowExperimentVariant[]>>({})
const loading = ref(true)
const saving = ref(false)
const mapperDialogOpen = ref(false)
const compatibilityDialogOpen = ref(false)
const experimentDialogOpen = ref(false)
const scopeDialogOpen = ref(false)
const variantDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<RoutingDeleteTarget | null>(null)

const mapperForm = reactive({
  name: 'Map public payload',
  mapperType: 'json',
  rulesJson: '{\n  "copy_all": true,\n  "mapping": {}\n}',
})

const compatibilityForm = reactive({
  versionId: '',
  inputSchemaId: '',
  mapperId: '',
  compatibilityType: 'native',
  isDefaultRoute: false,
  defaultValuesJson: '{}',
})

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

const editingMapperId = ref<number | null>(null)
const editingCompatibilityId = ref<number | null>(null)
const editingExperimentId = ref<number | null>(null)
const editingScopeId = ref<number | null>(null)
const editingVariantId = ref<number | null>(null)

const mapperBaseline = ref('')
const compatibilityBaseline = ref('')
const experimentBaseline = ref('')
const scopeBaseline = ref('')
const variantBaseline = ref('')

const activeVersions = computed(() => versions.value.filter(version => version.is_active && version.is_valid))
const activeSchemas = computed(() => schemas.value.filter(schema => schema.status === 'active'))

const isMapperDirty = computed(() => snapshotForm(mapperForm) !== mapperBaseline.value)
const isCompatibilityDirty = computed(() => snapshotForm(compatibilityForm) !== compatibilityBaseline.value)
const isExperimentDirty = computed(() => snapshotForm(experimentForm) !== experimentBaseline.value)
const isScopeDirty = computed(() => snapshotForm(scopeForm) !== scopeBaseline.value)
const isVariantDirty = computed(() => snapshotForm(variantForm) !== variantBaseline.value)
const dirtyForms = computed(() => [
  isMapperDirty.value ? t('workflowRouting.inputMappers') : '',
  isCompatibilityDirty.value ? t('workflowRouting.compatibilities') : '',
  isExperimentDirty.value ? t('workflowRouting.experiments') : '',
  isScopeDirty.value ? t('workflowRouting.scopeNumber', { id: editingScopeId.value || '' }) : '',
  isVariantDirty.value ? t('workflowRouting.variant') : '',
].filter(Boolean))
const hasUnsavedRoutingChanges = computed(() => dirtyForms.value.length > 0)

function snapshotForm(value: unknown) {
 return JSON.stringify(value)
}

function markMapperClean() {
  mapperBaseline.value = snapshotForm(mapperForm)
}

function markCompatibilityClean() {
  compatibilityBaseline.value = snapshotForm(compatibilityForm)
}

function markExperimentClean() {
  experimentBaseline.value = snapshotForm(experimentForm)
}

function markScopeClean() {
  scopeBaseline.value = snapshotForm(scopeForm)
}

function markVariantClean() {
  variantBaseline.value = snapshotForm(variantForm)
}

function markAllFormsClean() {
  markMapperClean()
  markCompatibilityClean()
  markExperimentClean()
  markScopeClean()
  markVariantClean()
}

function confirmDiscardRoutingChanges() {
  if (!hasUnsavedRoutingChanges.value) return true
  return window.confirm(t('workflowRouting.confirmDiscardChanges'))
}

function closeRoutingFormDialogs() {
  mapperDialogOpen.value = false
  compatibilityDialogOpen.value = false
  experimentDialogOpen.value = false
  scopeDialogOpen.value = false
  variantDialogOpen.value = false
}

function discardRoutingFormChanges() {
  cancelEditMapper()
  cancelEditCompatibility()
  cancelEditExperiment()
  cancelEditScope()
  cancelEditVariant()
  closeRoutingFormDialogs()
}

function prepareRoutingFormSwitch() {
  if (!confirmDiscardRoutingChanges()) return false
  discardRoutingFormChanges()
  return true
}

function versionLabel(versionId: number) {
  const version = versions.value.find(item => item.id === versionId)
  if (!version) return `#${versionId}`
  return `${version.name || 'Version'} v${version.version_number}`
}

function schemaLabel(schemaId: number) {
  const schema = schemas.value.find(item => item.id === schemaId)
  if (!schema) return `schema #${schemaId}`
  return `${schema.code} v${schema.version_number}`
}

function parseObjectJson(raw: string, label: string): Record<string, unknown> {
  const parsed = JSON.parse(raw || '{}')
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(`${label} must be a JSON object`)
  }
  return parsed as Record<string, unknown>
}

function validateMapperRules(rules: Record<string, unknown>) {
  if ('copy_all' in rules && typeof rules.copy_all !== 'boolean') {
    throw new Error('Mapper copy_all must be boolean')
  }
  if ('defaults' in rules && (!rules.defaults || typeof rules.defaults !== 'object' || Array.isArray(rules.defaults))) {
    throw new Error('Mapper defaults must be an object')
  }
  for (const key of ['fields', 'mapping', 'mappings']) {
    if (!(key in rules)) continue
    const mappings = rules[key]
    if (!mappings || typeof mappings !== 'object' || Array.isArray(mappings)) {
      throw new Error(`Mapper ${key} must be an object`)
    }
    for (const [targetPath, sourceSpec] of Object.entries(mappings as Record<string, unknown>)) {
      if (!targetPath.trim()) throw new Error('Mapper target path is required')
      if (typeof sourceSpec === 'string' && sourceSpec.trim()) continue
      if (sourceSpec && typeof sourceSpec === 'object' && !Array.isArray(sourceSpec)) {
        const spec = sourceSpec as Record<string, unknown>
        if ('literal' in spec) continue
        if (typeof spec.source === 'string' && spec.source.trim()) continue
        if (typeof spec.path === 'string' && spec.path.trim()) continue
      }
      throw new Error('Mapper source must be a path string or object with source, path, or literal')
    }
  }
}

async function loadData() {
  loading.value = true
  try {
    const [workflow, versionList, schemaList, mapperList, compatibilityList, experimentList] = await Promise.all([
      fetchWorkflow(workflowId.value),
      fetchVersionSummaries(workflowId.value),
      routing.fetchInputSchemas(workflowId.value),
      routing.fetchInputMappers(workflowId.value),
      routing.fetchCompatibilities(workflowId.value),
      routing.fetchExperiments(workflowId.value),
    ])
    workflowName.value = workflow.name
    versions.value = versionList
    schemas.value = schemaList
    mappers.value = mapperList
    compatibilities.value = compatibilityList
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

function startEditSchema(schema: WorkflowInputSchemaRecord) {
  if (!prepareRoutingFormSwitch()) return
  router.push(workflowInputSchemaEditorPath(orgId.value, workflowId.value, schema.id))
}

function openCreateSchemaDialog() {
  if (!prepareRoutingFormSwitch()) return
  router.push(workflowInputSchemaCreatePath(orgId.value, workflowId.value))
}

async function runRoutingAction(action: () => Promise<unknown>, successTitle: string, fallbackError: string) {
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

function openRoutingDeleteDialog(target: RoutingDeleteTarget) {
  deleteTarget.value = target
  deleteDialogOpen.value = true
}

function cancelRoutingDelete() {
  if (saving.value) return
  deleteDialogOpen.value = false
  deleteTarget.value = null
}

async function confirmRoutingDelete() {
  if (!deleteTarget.value) return
  const target = deleteTarget.value
  await runRoutingAction(target.action, target.successTitle, target.fallbackError)
  deleteDialogOpen.value = false
  deleteTarget.value = null
}

function onDeleteSchema(schema: WorkflowInputSchemaRecord) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteSchema', { name: `${schema.code} v${schema.version_number}` }),
    successTitle: t('workflowRouting.deletedInputSchema'),
    fallbackError: 'Failed to delete input schema',
    action: () => routing.deleteInputSchema(schema.id),
  })
}

function onSetDefaultSchema(schema: WorkflowInputSchemaRecord) {
  routing.setInputSchemaDefault(schema.id).then(loadData)
}

function onActivateSchema(schema: WorkflowInputSchemaRecord) {
  routing.updateInputSchemaStatus(schema.id, 'active').then(loadData)
}

async function onCreateMapper() {
  saving.value = true
  try {
    const rules = parseObjectJson(mapperForm.rulesJson, 'Mapper rules')
    validateMapperRules(rules)
    if (editingMapperId.value) {
      await routing.updateInputMapper(editingMapperId.value, {
        name: mapperForm.name.trim(),
        mapper_type: mapperForm.mapperType,
        rules,
      })
      editingMapperId.value = null
      toast({ title: t('workflowRouting.updatedMapper') })
    }
    else {
      await routing.createInputMapper(workflowId.value, {
        name: mapperForm.name.trim(),
        mapper_type: mapperForm.mapperType,
        rules,
        is_active: true,
      })
      toast({ title: t('workflowRouting.createdMapper') })
    }
    mapperDialogOpen.value = false
    cancelEditMapper()
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateMapper')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function startEditMapper(mapper: WorkflowInputMapperRecord) {
  if (!prepareRoutingFormSwitch()) return
  editingMapperId.value = mapper.id
  mapperForm.name = mapper.name
  mapperForm.mapperType = mapper.mapper_type
  mapperForm.rulesJson = JSON.stringify(mapper.rules ?? {}, null, 2)
  markMapperClean()
  mapperDialogOpen.value = true
}

function cancelEditMapper() {
  editingMapperId.value = null
  mapperForm.name = 'Map public payload'
  mapperForm.mapperType = 'json'
  mapperForm.rulesJson = '{\n  "copy_all": true,\n  "mapping": {}\n}'
  markMapperClean()
}

function openCreateMapperDialog() {
  if (!prepareRoutingFormSwitch()) return
  cancelEditMapper()
  mapperDialogOpen.value = true
}

async function onToggleMapper(mapper: WorkflowInputMapperRecord) {
  await runRoutingAction(
    () => routing.setInputMapperActive(mapper.id, !mapper.is_active),
    mapper.is_active ? t('workflowRouting.deactivatedMapper') : t('workflowRouting.activatedMapper'),
    'Failed to update mapper',
  )
}

function onDeleteMapper(mapper: WorkflowInputMapperRecord) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteMapper', { name: mapper.name }),
    successTitle: t('workflowRouting.deletedMapper'),
    fallbackError: 'Failed to delete mapper',
    action: () => routing.deleteInputMapper(mapper.id),
  })
}

async function onCreateCompatibility() {
  saving.value = true
  try {
    const payload = {
      workflow_input_schema_id: Number(compatibilityForm.inputSchemaId),
      workflow_input_mapper_id: compatibilityForm.mapperId ? Number(compatibilityForm.mapperId) : undefined,
      compatibility_type: compatibilityForm.compatibilityType,
      default_values: parseObjectJson(compatibilityForm.defaultValuesJson, 'Default values'),
      is_active: true,
      is_default_route: compatibilityForm.isDefaultRoute,
    }
    if (editingCompatibilityId.value) {
      await routing.updateCompatibility(editingCompatibilityId.value, payload)
      editingCompatibilityId.value = null
      toast({ title: t('workflowRouting.updatedCompatibility') })
    }
    else {
      await routing.createCompatibility(Number(compatibilityForm.versionId), payload)
      toast({ title: t('workflowRouting.createdCompatibility') })
    }
    compatibilityDialogOpen.value = false
    cancelEditCompatibility()
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateCompatibility')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function startEditCompatibility(compatibility: WorkflowSchemaCompatibility) {
  if (!prepareRoutingFormSwitch()) return
  editingCompatibilityId.value = compatibility.id
  compatibilityForm.versionId = String(compatibility.workflow_version_id)
  compatibilityForm.inputSchemaId = String(compatibility.workflow_input_schema_id)
  compatibilityForm.mapperId = compatibility.workflow_input_mapper_id ? String(compatibility.workflow_input_mapper_id) : ''
  compatibilityForm.compatibilityType = compatibility.compatibility_type
  compatibilityForm.isDefaultRoute = compatibility.is_default_route
  compatibilityForm.defaultValuesJson = JSON.stringify(compatibility.default_values ?? {}, null, 2)
  markCompatibilityClean()
  compatibilityDialogOpen.value = true
}

function cancelEditCompatibility() {
  editingCompatibilityId.value = null
  compatibilityForm.versionId = ''
  compatibilityForm.inputSchemaId = ''
  compatibilityForm.mapperId = ''
  compatibilityForm.compatibilityType = 'native'
  compatibilityForm.isDefaultRoute = false
  compatibilityForm.defaultValuesJson = '{}'
  markCompatibilityClean()
}

function openCreateCompatibilityDialog() {
  if (!prepareRoutingFormSwitch()) return
  cancelEditCompatibility()
  compatibilityDialogOpen.value = true
}

async function onDeactivateCompatibility(compatibility: WorkflowSchemaCompatibility) {
  await runRoutingAction(
    () => routing.deactivateCompatibility(compatibility.id),
    t('workflowRouting.deactivatedCompatibility'),
    'Failed to deactivate compatibility',
  )
}

function onDeleteCompatibility(compatibility: WorkflowSchemaCompatibility) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteCompatibility', { route: `${schemaLabel(compatibility.workflow_input_schema_id)} ${t('workflowRouting.to')} ${versionLabel(compatibility.workflow_version_id)}` }),
    successTitle: t('workflowRouting.deletedCompatibility'),
    fallbackError: 'Failed to delete compatibility',
    action: () => routing.deleteCompatibility(compatibility.id),
  })
}

function onMakeDefaultCompatibility(compatibility: WorkflowSchemaCompatibility) {
  routing.setCompatibilityDefaultRoute(compatibility.id, true).then(loadData)
}

async function onCreateExperiment() {
  saving.value = true
  try {
    if (editingExperimentId.value) {
      await routing.updateExperiment(editingExperimentId.value, {
        name: experimentForm.name.trim(),
        experiment_type: experimentForm.experimentType,
      })
      await routing.updateExperimentStatus(editingExperimentId.value, experimentForm.status)
      editingExperimentId.value = null
      toast({ title: t('workflowRouting.updatedExperiment') })
    }
    else {
      await routing.createExperiment(workflowId.value, {
        name: experimentForm.name.trim(),
        experiment_type: experimentForm.experimentType,
        status: experimentForm.status,
      })
      toast({ title: t('workflowRouting.createdExperiment') })
    }
    experimentDialogOpen.value = false
    cancelEditExperiment()
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateExperiment')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function startEditExperiment(experiment: WorkflowExperiment) {
  if (!prepareRoutingFormSwitch()) return
  editingExperimentId.value = experiment.id
  experimentForm.name = experiment.name
  experimentForm.experimentType = experiment.experiment_type
  experimentForm.status = experiment.status
  markExperimentClean()
  experimentDialogOpen.value = true
}

function cancelEditExperiment() {
  editingExperimentId.value = null
  experimentForm.name = 'Canary route'
  experimentForm.experimentType = 'canary'
  experimentForm.status = 'draft'
  markExperimentClean()
}

function openCreateExperimentDialog() {
  if (!prepareRoutingFormSwitch()) return
  cancelEditExperiment()
  experimentDialogOpen.value = true
}

async function onPauseExperiment(experiment: WorkflowExperiment) {
  await runRoutingAction(
    () => routing.updateExperimentStatus(experiment.id, 'paused'),
    t('workflowRouting.pausedExperiment'),
    'Failed to pause experiment',
  )
}

function onDeleteExperiment(experiment: WorkflowExperiment) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteExperiment', { name: experiment.name }),
    successTitle: t('workflowRouting.deletedExperiment'),
    fallbackError: 'Failed to delete experiment',
    action: () => routing.deleteExperiment(experiment.id),
  })
}

function onActivateExperiment(experiment: WorkflowExperiment) {
  routing.updateExperimentStatus(experiment.id, 'active').then(loadData)
}

async function onCreateScope() {
  saving.value = true
  try {
    const payload = {
      workflow_input_schema_id: Number(scopeForm.inputSchemaId),
      traffic_percent: Number(scopeForm.trafficPercent),
      fallback_policy: scopeForm.fallbackPolicy,
      fallback_workflow_version_id: scopeForm.fallbackVersionId ? Number(scopeForm.fallbackVersionId) : undefined,
      traffic_conditions: parseObjectJson(scopeForm.trafficConditionsJson, 'Traffic conditions'),
    }
    if (editingScopeId.value) {
      await routing.updateExperimentScope(editingScopeId.value, payload)
      editingScopeId.value = null
      toast({ title: t('workflowRouting.updatedScope') })
    }
    else {
      await routing.createExperimentScope(Number(scopeForm.experimentId), payload)
      toast({ title: t('workflowRouting.createdScope') })
    }
    scopeDialogOpen.value = false
    cancelEditScope()
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateScope')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function startEditScope(experiment: WorkflowExperiment, scope: WorkflowExperimentScope) {
  if (!prepareRoutingFormSwitch()) return
  editingScopeId.value = scope.id
  scopeForm.experimentId = String(experiment.id)
  scopeForm.inputSchemaId = String(scope.workflow_input_schema_id)
  scopeForm.trafficPercent = scope.traffic_percent
  scopeForm.fallbackPolicy = scope.fallback_policy
  scopeForm.fallbackVersionId = scope.fallback_workflow_version_id ? String(scope.fallback_workflow_version_id) : ''
  scopeForm.trafficConditionsJson = JSON.stringify(scope.traffic_conditions ?? {}, null, 2)
  markScopeClean()
  scopeDialogOpen.value = true
}

function cancelEditScope() {
  editingScopeId.value = null
  scopeForm.experimentId = ''
  scopeForm.inputSchemaId = ''
  scopeForm.trafficPercent = 100
  scopeForm.fallbackPolicy = 'default_route'
  scopeForm.fallbackVersionId = ''
  scopeForm.trafficConditionsJson = '{}'
  markScopeClean()
}

function openCreateScopeDialog(experiment?: WorkflowExperiment) {
  if (!prepareRoutingFormSwitch()) return
  cancelEditScope()
  if (experiment) scopeForm.experimentId = String(experiment.id)
  markScopeClean()
  scopeDialogOpen.value = true
}

function onDeleteScope(scope: WorkflowExperimentScope) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteScope', { id: scope.id }),
    successTitle: t('workflowRouting.deletedScope'),
    fallbackError: 'Failed to delete scope',
    action: () => routing.deleteExperimentScope(scope.id),
  })
}

async function onCreateVariant() {
  saving.value = true
  try {
    const payload = {
      workflow_version_id: Number(variantForm.versionId),
      traffic_weight: Number(variantForm.trafficWeight),
      is_control_group: variantForm.isControlGroup,
      is_active: variantForm.isActive,
    }
    if (editingVariantId.value) {
      await routing.updateExperimentVariant(editingVariantId.value, payload)
      editingVariantId.value = null
      toast({ title: t('workflowRouting.updatedVariant') })
    }
    else {
      await routing.createExperimentVariant(Number(variantForm.scopeId), payload)
      toast({ title: t('workflowRouting.createdVariant') })
    }
    variantDialogOpen.value = false
    cancelEditVariant()
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorCreateVariant')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function startEditVariant(scope: WorkflowExperimentScope, variant: WorkflowExperimentVariant) {
  if (!prepareRoutingFormSwitch()) return
  editingVariantId.value = variant.id
  variantForm.scopeId = String(scope.id)
  variantForm.versionId = String(variant.workflow_version_id)
  variantForm.trafficWeight = variant.traffic_weight
  variantForm.isControlGroup = variant.is_control_group
  variantForm.isActive = variant.is_active
  markVariantClean()
  variantDialogOpen.value = true
}

function cancelEditVariant() {
  editingVariantId.value = null
  variantForm.scopeId = ''
  variantForm.versionId = ''
  variantForm.trafficWeight = 100
  variantForm.isControlGroup = false
  variantForm.isActive = true
  markVariantClean()
}

function openCreateVariantDialog(scope?: WorkflowExperimentScope) {
  if (!prepareRoutingFormSwitch()) return
  cancelEditVariant()
  if (scope) variantForm.scopeId = String(scope.id)
  markVariantClean()
  variantDialogOpen.value = true
}

function onDeleteVariant(variant: WorkflowExperimentVariant) {
  openRoutingDeleteDialog({
    message: t('workflowRouting.confirmDeleteVariant', { id: variant.id }),
    successTitle: t('workflowRouting.deletedVariant'),
    fallbackError: 'Failed to delete variant',
    action: () => routing.deleteExperimentVariant(variant.id),
  })
}

function handleBeforeUnload(event: BeforeUnloadEvent) {
  if (!hasUnsavedRoutingChanges.value) return
  event.preventDefault()
  event.returnValue = ''
}

markAllFormsClean()

onMounted(() => {
  void loadData()
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload)
})

onBeforeRouteLeave(() => {
  if (confirmDiscardRoutingChanges()) return
  return false
})
</script>

<template>
  <div class="space-y-6">
    <div class="relative overflow-hidden rounded-3xl border bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.18),_transparent_34%),linear-gradient(135deg,_hsl(var(--background)),_hsl(var(--muted)))] p-6">
      <div class="relative z-10 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <Button variant="ghost" class="-ml-3 mb-3 gap-2" @click="router.push(`/org/${orgId}/workflows/${workflowId}`)">
            <ArrowLeft class="h-4 w-4" />
            {{ t('common.back') }}
          </Button>
          <div class="flex items-center gap-3">
            <div class="rounded-2xl bg-primary/10 p-3 text-primary">
              <Route class="h-6 w-6" />
            </div>
            <div>
              <h1 class="text-2xl font-semibold">{{ t('workflowRouting.title') }}</h1>
              <p class="text-sm text-muted-foreground">
                {{ workflowName || t('common.loading') }}
              </p>
            </div>
          </div>
        </div>
        <div class="grid grid-cols-3 gap-3 text-center text-sm">
          <div class="rounded-2xl border bg-background/70 p-3">
            <div class="text-xl font-semibold">{{ schemas.length }}</div>
            <div class="text-muted-foreground">{{ t('workflowRouting.schemas') }}</div>
          </div>
          <div class="rounded-2xl border bg-background/70 p-3">
            <div class="text-xl font-semibold">{{ compatibilities.length }}</div>
            <div class="text-muted-foreground">{{ t('workflowRouting.routes') }}</div>
          </div>
          <div class="rounded-2xl border bg-background/70 p-3">
            <div class="text-xl font-semibold">{{ experiments.length }}</div>
            <div class="text-muted-foreground">{{ t('workflowRouting.experiments') }}</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>

    <div v-else class="space-y-6">
      <div
        v-if="hasUnsavedRoutingChanges"
        class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-100"
      >
        <div>
          <p class="font-medium">{{ t('workflowRouting.unsavedChanges') }}</p>
          <p class="text-xs opacity-80">
            {{ t('workflowRouting.unsavedChangesDescription', { forms: dirtyForms.join(', ') }) }}
          </p>
        </div>
        <Button variant="outline" size="sm" :disabled="saving" @click="discardRoutingFormChanges">
          {{ t('workflowRouting.discardChanges') }}
        </Button>
      </div>
      <div class="space-y-6">
        <RoutingInputSchemasList
          :schemas="schemas"
          :saving="saving"
          @create="openCreateSchemaDialog"
          @edit="startEditSchema"
          @set-default="onSetDefaultSchema"
          @activate="onActivateSchema"
          @delete="onDeleteSchema"
        />

        <RoutingInputMappersList
          :mappers="mappers"
          :saving="saving"
          @create="openCreateMapperDialog"
          @edit="startEditMapper"
          @toggle="onToggleMapper"
          @delete="onDeleteMapper"
        />

        <RoutingCompatibilitiesList
          :compatibilities="compatibilities"
          :saving="saving"
          :schema-label="schemaLabel"
          :version-label="versionLabel"
          @create="openCreateCompatibilityDialog"
          @edit="startEditCompatibility"
          @make-default="onMakeDefaultCompatibility"
          @deactivate="onDeactivateCompatibility"
          @delete="onDeleteCompatibility"
        />

        <RoutingExperimentsList
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
      </div>
    </div>

    <RoutingInputMapperDialog
      v-model:open="mapperDialogOpen"
      :form="mapperForm"
      :saving="saving"
      :editing-id="editingMapperId"
      @submit="onCreateMapper"
      @cancel="cancelEditMapper"
    />

    <RoutingCompatibilityDialog
      v-model:open="compatibilityDialogOpen"
      :form="compatibilityForm"
      :active-versions="activeVersions"
      :active-schemas="activeSchemas"
      :mappers="mappers"
      :saving="saving"
      :editing-id="editingCompatibilityId"
      :version-label="versionLabel"
      :schema-label="schemaLabel"
      @submit="onCreateCompatibility"
      @cancel="cancelEditCompatibility"
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
      @cancel="cancelRoutingDelete"
      @confirm="confirmRoutingDelete"
    />
  </div>
</template>


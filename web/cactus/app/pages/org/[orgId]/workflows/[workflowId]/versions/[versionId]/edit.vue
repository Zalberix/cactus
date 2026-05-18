<script setup lang="ts">
import { Save, Play, Pause, AlertCircle, Copy, FileJson } from 'lucide-vue-next'
import type { Connection } from '@vue-flow/core'
import type { VersionSummary } from '~/composables/useVersions'
import type { WorkType } from '~/composables/useWorkers'
import type { WorkflowInputSchemaField } from '~/composables/useWorkflows'
import type { StepData } from '~/composables/useDagEditor'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import StepToolbar from '~/components/dag/StepToolbar.vue'
import StepPanel from '~/components/dag/StepPanel.vue'
import StepSchemaChoiceDialog from '~/components/dag/StepSchemaChoiceDialog.vue'
import WorkflowSchemaDialog from '~/components/dag/WorkflowSchemaDialog.vue'
import NodeEditor from '~/components/dag/node-editor/NodeEditor.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import { editorSurfaceForStep, isVersionReadOnly } from '~/components/dag/editor-utils'
import {
  schemaChoiceOptions,
  shouldPromptForSchemaChoice,
} from '~/components/dag/step-toolbar-utils'
import type { StepAddPayload } from '~/components/dag/step-toolbar-utils'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))
const routeVersionId = computed(() => Number(route.params.versionId || 0))

const { fetchWorkflow, fetchWorkflowInputSchema, upsertWorkflowInputSchemaField, deleteWorkflowInputSchemaField } = useWorkflows()
const {
  fetchVersionSummaries,
  copyVersion,
  updateVersionName,
  activateVersion,
  deactivateVersion,
  fetchWorkTypes,
} = useVersions()

// Workflow state
const workflowName = ref('')
const versions = ref<VersionSummary[]>([])
const selectedVersionId = ref<number | null>(null)
const workTypes = ref<WorkType[]>([])
const pageLoading = ref(true)
const saving = ref(false)
const schemaOpen = ref(false)
const schemaSaving = ref(false)
const inputSchema = ref<Record<string, unknown> | null>(null)
const deactivateOpen = ref(false)
const schemaChoiceOpen = ref(false)
const pendingStepPayload = ref<StepAddPayload | null>(null)
const showValidationDialog = ref(false)

const currentVersion = computed(() =>
  versions.value.find(v => v.id === selectedVersionId.value),
)

const currentVersionName = computed({
  get: () => currentVersion.value?.name || t('editor.versionNumber', { number: currentVersion.value?.version_number ?? '' }),
  set: (value: string) => {
    const version = currentVersion.value
    if (version) version.name = value
  },
})

const isVersionActive = computed(() => currentVersion.value?.is_active ?? false)
const isCurrentVersionReadOnly = computed(() => isVersionReadOnly(currentVersion.value))

// DAG editor composable
const dagEditor = useDagEditor(workflowId, selectedVersionId, isCurrentVersionReadOnly)

// Node editor composable
const nodeEditor = useNodeEditor()

const canActivate = computed(() =>
  Boolean(selectedVersionId.value && currentVersion.value?.is_valid && !dagEditor.isDirty.value),
)

const canvasEdges = computed(() =>
  dagEditor.edges.value.map(edge => ({
    ...edge,
    selected: edge.id === dagEditor.selectedEdgeId.value,
  })),
)

// Load workflow data on mount
async function loadAll() {
  pageLoading.value = true
  try {
    const [wf, vers, wts] = await Promise.all([
      fetchWorkflow(workflowId.value),
      fetchVersionSummaries(workflowId.value),
      fetchWorkTypes(),
    ])

    workflowName.value = wf.name
    versions.value = vers
    workTypes.value = wts

    // Auto-select: query version, else active version, else latest
    if (vers.length > 0) {
      const routeVersion = vers.find(v => v.id === routeVersionId.value)
      const active = vers.find(v => v.is_active)
      const latest = vers[vers.length - 1]
      selectedVersionId.value = (routeVersion ?? active ?? latest).id
    }
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    pageLoading.value = false
  }
}

// When version changes, load its steps
watch(selectedVersionId, async (vid) => {
  showValidationDialog.value = false
  if (vid) {
    try {
      await dagEditor.loadSteps(vid)
    }
    catch (err) {
      toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
    }
  }
})

watch(isCurrentVersionReadOnly, (readOnly) => {
  if (readOnly) {
    nodeEditor.close()
  }
})

async function onSave() {
  if (isCurrentVersionReadOnly.value) {
    await onCreateEditableCopy()
    return
  }

  saving.value = true
  try {
    const success = await dagEditor.saveVersion()
    if (success) {
      toast({ title: t('editor.validated') })
      versions.value = await fetchVersionSummaries(workflowId.value)
      if (selectedVersionId.value) {
        inputSchema.value = await fetchWorkflowInputSchema(workflowId.value)
      }
    }
    else if (dagEditor.validationErrors.value.length > 0) {
      showValidationDialog.value = true
      toast({ title: t('error.dagValidation'), variant: 'destructive' })
    }
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

async function onActivate() {
  if (!selectedVersionId.value || !canActivate.value) return
  try {
    await activateVersion(selectedVersionId.value)
    toast({ title: t('editor.activateVersion') })
    versions.value = await fetchVersionSummaries(workflowId.value)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onDeactivate() {
  if (!selectedVersionId.value) return
  try {
    await deactivateVersion(selectedVersionId.value)
    deactivateOpen.value = false
    toast({ title: t('editor.deactivateVersion') })
    versions.value = await fetchVersionSummaries(workflowId.value)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onCreateEditableCopy() {
  if (!selectedVersionId.value) return
  try {
    const ver = await copyVersion(selectedVersionId.value)
    versions.value = await fetchVersionSummaries(workflowId.value)
    selectedVersionId.value = ver.id
    await router.replace(workflowVersionEditorPath(orgId.value, workflowId.value, ver.id))
    toast({ title: t('editor.createEditableCopy') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onVersionNameBlur() {
  const version = currentVersion.value
  if (!version || isCurrentVersionReadOnly.value) return
  try {
    await updateVersionName(version.id, currentVersionName.value)
    versions.value = await fetchVersionSummaries(workflowId.value)
  }
  catch {
    // Non-critical inline rename.
  }
}

async function openSchemaDialog() {
  if (!currentVersion.value) return
  try {
    inputSchema.value = await fetchWorkflowInputSchema(workflowId.value)
    schemaOpen.value = true
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onSchemaSaveField(field: WorkflowInputSchemaField) {
  schemaSaving.value = true
  try {
    inputSchema.value = await upsertWorkflowInputSchemaField(workflowId.value, field)
    versions.value = await fetchVersionSummaries(workflowId.value)
    toast({ title: t('workflowInputs.saved') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    schemaSaving.value = false
  }
}

async function onSchemaDeleteField(field: { name: string }) {
  schemaSaving.value = true
  try {
    inputSchema.value = await deleteWorkflowInputSchemaField(workflowId.value, field.name)
    versions.value = await fetchVersionSummaries(workflowId.value)
    toast({ title: t('workflowInputs.deleted') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    schemaSaving.value = false
  }
}

async function onWorkflowInputsChanged() {
  try {
    versions.value = await fetchVersionSummaries(workflowId.value)
    if (schemaOpen.value) {
      inputSchema.value = await fetchWorkflowInputSchema(workflowId.value)
    }
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

function onConnect(params: Connection) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.connectSteps(params)
}

function onNodeDragStop(nodeId: string, position: { x: number; y: number }) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.onNodeDragStop(nodeId, position)
}

function onNodeClick(nodeId: string) {
  const node = dagEditor.nodes.value.find(n => n.id === nodeId)
  const surface = editorSurfaceForStep(node?.data ?? {})
  dagEditor.selectNode(surface === 'none' ? null : nodeId)
}

function onNodeDoubleClick(nodeId: string) {
  if (isCurrentVersionReadOnly.value) return
  const node = dagEditor.nodes.value.find(n => n.id === nodeId)
  const surface = editorSurfaceForStep(node?.data ?? {})
  if (surface === 'node-editor') {
    nodeEditor.open(nodeId)
    return
  }
  if (surface === 'step-panel') {
    dagEditor.selectNode(nodeId)
  }
}

function onEdgeClick(edgeId: string) {
  dagEditor.selectEdge(edgeId)
}

function onRemoveEdge(edgeId: string) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.removeEdge(edgeId)
}

function addStepFromPayload(payload: StepAddPayload) {
  if (isCurrentVersionReadOnly.value) return

  if (shouldPromptForSchemaChoice(payload.stepType, payload.schemas, payload.workerSettingsSchemaId)) {
    pendingStepPayload.value = payload
    schemaChoiceOpen.value = true
    return
  }

  dagEditor.addStep(
    payload.stepType,
    payload.workTypeId,
    payload.workTypeCode,
    payload.workerSettingsSchemaId,
    payload.position,
    payload.name,
  )
}

function onDrop(payload: StepAddPayload) {
  addStepFromPayload(payload)
}

function onToolbarAddStep(payload: StepAddPayload) {
  addStepFromPayload(payload)
}

function onChooseStepSchema(schemaId: number) {
  const payload = pendingStepPayload.value
  if (!payload) return

  schemaChoiceOpen.value = false
  pendingStepPayload.value = null
  addStepFromPayload({
    ...payload,
    workerSettingsSchemaId: schemaId,
    schemas: schemaChoiceOptions(payload.schemas),
  })
}

function onDeleteSelected() {
  if (isCurrentVersionReadOnly.value) return
  if (dagEditor.selectedEdgeId.value) {
    dagEditor.removeEdge(dagEditor.selectedEdgeId.value)
    return
  }

  if (dagEditor.selectedNodeId.value) {
    dagEditor.removeStep(dagEditor.selectedNodeId.value)
  }
}

function onPanelUpdateData(nodeId: string, data: Partial<StepData>) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.updateNodeData(nodeId, data)
}

function onPanelUpdateStep(stepId: string, data: Record<string, unknown>) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.updateStepOnServer(stepId, data)
}

function onPanelDeleteStep(stepId: string) {
  if (isCurrentVersionReadOnly.value) return
  dagEditor.removeStep(stepId)
}

function onPanelOpenEditor(nodeId: string) {
  if (isCurrentVersionReadOnly.value) return
  const node = dagEditor.nodes.value.find(n => n.id === nodeId)
  if (node?.data.controlKind === 'start') return
  nodeEditor.open(nodeId)
}

async function onNodeEditorSaveSettings(nodeId: string, settingsData: Record<string, unknown>) {
  if (isCurrentVersionReadOnly.value) return
  try {
    await dagEditor.updateTaskSettingsOnServer(nodeId, settingsData)
    dagEditor.updateNodeData(nodeId, { config: settingsData })
    toast({ title: t('nodeEditor.settingsSaved') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onNodeEditorSaveInputMapping(nodeId: string, inputMapping: Array<{ target: string, source: string }>) {
  if (isCurrentVersionReadOnly.value) return
  try {
    await dagEditor.updateTaskInputMappingOnServer(nodeId, inputMapping)
    const inputMappingRecord = Object.fromEntries(inputMapping.map(entry => [entry.target, entry.source]))
    dagEditor.updateNodeData(nodeId, { inputMapping: inputMappingRecord })
    toast({ title: t('nodeEditor.inputMappingSaved') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

function dismissErrors() {
  dagEditor.validationErrors.value = []
  showValidationDialog.value = false
}

function onValidationDialogUpdate(open: boolean) {
  if (!open) dismissErrors()
}

onMounted(() => {
  loadAll()
})
</script>

<template>
  <div class="flex h-[calc(100vh-3.5rem)] flex-col">
    <!-- Header bar -->
    <div class="flex items-center gap-3 border-b px-4 py-2">
      <div class="min-w-0 space-y-1">
        <div class="h-7 truncate text-sm font-semibold leading-7">
          {{ workflowName || t('common.loading') }}
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <Input
            v-model="currentVersionName"
            :readonly="!currentVersion || isCurrentVersionReadOnly"
            class="h-7 w-56 border-0 p-0 text-xs font-medium text-muted-foreground shadow-none focus-visible:ring-0"
            @blur="onVersionNameBlur"
            @keydown.enter="($event.target as HTMLInputElement)?.blur()"
          />
          <Badge v-if="currentVersion?.is_active" class="h-5 px-1.5 text-[10px]">
            {{ t('editor.activeVersion') }}
          </Badge>
          <Badge
            v-if="currentVersion?.is_valid"
            variant="secondary"
            class="h-5 px-1.5 text-[10px] bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300"
          >
            {{ t('editor.valid') }}
          </Badge>
          <Badge
            v-else-if="currentVersion"
            variant="secondary"
            class="h-5 px-1.5 text-[10px] bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
          >
            {{ t('editor.invalid') }}
          </Badge>
          <Badge
            v-if="dagEditor.isDirty.value"
            variant="outline"
            class="h-5 px-1.5 text-[10px]"
          >
            {{ t('editor.validationRequired') }}
          </Badge>
          <Badge
            v-if="isCurrentVersionReadOnly"
            variant="outline"
            class="h-5 px-1.5 text-[10px]"
          >
            {{ t('editor.readOnly') }}
          </Badge>
        </div>
      </div>

      <div class="flex-1" />

      <Button
        size="sm"
        variant="outline"
        class="gap-2"
        :disabled="!currentVersion"
        @click="openSchemaDialog"
      >
        <FileJson class="h-4 w-4" />
        {{ t('editor.inputSchema') }}
      </Button>

      <Button
        size="sm"
        :disabled="saving || !selectedVersionId"
        @click="onSave"
      >
        <Copy v-if="isCurrentVersionReadOnly" class="mr-2 h-4 w-4" />
        <Save v-else class="mr-2 h-4 w-4" />
        {{ isCurrentVersionReadOnly ? t('editor.createEditableCopy') : t('editor.validate') }}
      </Button>

      <Button
        v-if="!isVersionActive"
        size="sm"
        variant="outline"
        :disabled="!canActivate"
        @click="onActivate"
      >
        <Play class="mr-2 h-4 w-4" />
        {{ t('editor.activateVersion') }}
      </Button>
      <Button
        v-else
        size="sm"
        variant="outline"
        :disabled="!selectedVersionId"
        @click="deactivateOpen = true"
      >
        <Pause class="mr-2 h-4 w-4" />
        {{ t('editor.deactivateVersion') }}
      </Button>
    </div>

    <Dialog v-model:open="showValidationDialog" @update:open="onValidationDialogUpdate">
      <DialogContent class="max-w-2xl">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <AlertCircle class="h-4 w-4 text-red-500" />
            {{ t('editor.validationErrors') }}
          </DialogTitle>
          <DialogDescription>
            {{ t('error.dagValidation') }}
          </DialogDescription>
        </DialogHeader>
        <div class="max-h-72 overflow-y-auto pl-5">
          <ul class="space-y-1 list-disc text-sm">
            <li v-for="(error, i) in dagEditor.validationErrors.value" :key="i">
              {{ error.message }}
            </li>
          </ul>
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            @click="dismissErrors"
          >
            {{ t('destructive.cancel') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Editor layout -->
    <div v-if="!pageLoading && versions.length > 0" class="flex flex-1 overflow-hidden">
      <!-- Left toolbar -->
      <StepToolbar :disabled="isCurrentVersionReadOnly" @add-step="onToolbarAddStep" />

      <!-- Center canvas -->
      <div class="flex-1">
        <DagCanvas
          :mode="isCurrentVersionReadOnly ? 'view' : 'edit'"
          :nodes="dagEditor.nodes.value"
          :edges="canvasEdges"
          @connect="onConnect"
          @node-drag-stop="onNodeDragStop"
          @node-click="onNodeClick"
          @node-double-click="onNodeDoubleClick"
          @edge-click="onEdgeClick"
          @remove-edge="onRemoveEdge"
          @drop="onDrop"
          @delete-selected="onDeleteSelected"
        />
      </div>

      <!-- Right panel -->
      <StepPanel
        v-if="dagEditor.selectedNode.value && editorSurfaceForStep(dagEditor.selectedNode.value.data) !== 'node-editor'"
        :node="dagEditor.selectedNode.value"
        :work-types="workTypes"
        :read-only="isCurrentVersionReadOnly"
        @close="dagEditor.selectNode(null)"
        @open-editor="onPanelOpenEditor"
        @update-data="onPanelUpdateData"
        @update-step="onPanelUpdateStep"
        @delete-step="onPanelDeleteStep"
      />
    </div>

    <!-- No versions state -->
    <div v-else-if="!pageLoading && versions.length === 0" class="flex flex-1 items-center justify-center">
      <EmptyState
        :heading="t('empty.versions.heading')"
        :body="t('empty.versions.body')"
      />
    </div>

    <!-- Loading state -->
    <div v-else class="flex flex-1 items-center justify-center">
      <p class="text-muted-foreground">{{ t('common.loading') }}</p>
    </div>

    <!-- Node Editor (full-screen sheet) -->
    <NodeEditor
      v-model:open="nodeEditor.isOpen.value"
      :node-id="nodeEditor.editingNodeId.value"
      :workflow-id="workflowId"
      :version-id="selectedVersionId"
      :all-nodes="dagEditor.nodes.value"
      :all-edges="dagEditor.edges.value"
      @save-settings="onNodeEditorSaveSettings"
      @save-input-mapping="onNodeEditorSaveInputMapping"
      @workflow-inputs-changed="onWorkflowInputsChanged"
    />

    <WorkflowSchemaDialog
      v-model:open="schemaOpen"
      :schema="inputSchema"
      :saving="schemaSaving"
      @save-field="onSchemaSaveField"
      @delete-field="onSchemaDeleteField"
    />

    <StepSchemaChoiceDialog
      v-model:open="schemaChoiceOpen"
      :step-name="pendingStepPayload?.name"
      :schemas="schemaChoiceOptions(pendingStepPayload?.schemas)"
      @choose="onChooseStepSchema"
    />

    <!-- Deactivate Confirmation Dialog -->
    <Dialog v-model:open="deactivateOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('destructive.deactivateVersion.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deactivateVersion.body', { number: currentVersion?.version_number ?? '' }) }}
            {{ t('destructive.deactivateVersion.trafficBody') }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" @click="deactivateOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button
            variant="destructive"
            @click="onDeactivate"
          >
            {{ t('destructive.deactivateVersion.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

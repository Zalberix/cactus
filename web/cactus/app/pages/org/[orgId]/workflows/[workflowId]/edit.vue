<script setup lang="ts">
import '@vue-flow/core/dist/style.css'
import { Save, Play, Pause, AlertCircle, X } from 'lucide-vue-next'
import type { Connection } from '@vue-flow/core'
import type { Version, WorkType } from '~/composables/useVersions'
import type { StepData } from '~/composables/useDagEditor'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import StepToolbar from '~/components/dag/StepToolbar.vue'
import StepPanel from '~/components/dag/StepPanel.vue'
import VersionSelector from '~/components/dag/VersionSelector.vue'
import NodeEditor from '~/components/dag/node-editor/NodeEditor.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
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

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const { fetchWorkflow, updateWorkflow } = useWorkflows()
const {
  fetchVersions,
  createVersion,
  activateVersion,
  deactivateVersion,
  fetchWorkTypes,
} = useVersions()

// Workflow state
const workflowName = ref('')
const versions = ref<Version[]>([])
const selectedVersionId = ref<number | null>(null)
const workTypes = ref<WorkType[]>([])
const pageLoading = ref(true)
const saving = ref(false)
const deactivateOpen = ref(false)

// DAG editor composable
const dagEditor = useDagEditor(workflowId, selectedVersionId)

// Node editor composable
const nodeEditor = useNodeEditor()

const currentVersion = computed(() =>
  versions.value.find(v => v.id === selectedVersionId.value),
)

const isVersionActive = computed(() => currentVersion.value?.is_active ?? false)

// Load workflow data on mount
async function loadAll() {
  pageLoading.value = true
  try {
    const [wf, vers, wts] = await Promise.all([
      fetchWorkflow(workflowId.value),
      fetchVersions(workflowId.value),
      fetchWorkTypes(),
    ])

    workflowName.value = wf.name
    versions.value = vers
    workTypes.value = wts

    // Auto-select: prefer active version, else latest
    if (vers.length > 0) {
      const active = vers.find(v => v.is_active)
      const latest = vers[vers.length - 1]
      selectedVersionId.value = (active ?? latest).id
    }
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    pageLoading.value = false
  }
}

// When version changes, load its steps
watch(selectedVersionId, async (vid) => {
  if (vid) {
    try {
      await dagEditor.loadSteps(vid)
    }
    catch {
      toast({ title: t('error.server'), variant: 'destructive' })
    }
  }
})

async function onSave() {
  saving.value = true
  try {
    const success = await dagEditor.saveVersion()
    if (success) {
      toast({ title: t('editor.validated') })
      versions.value = await fetchVersions(workflowId.value)
    }
    else if (dagEditor.validationErrors.value.length > 0) {
      toast({ title: t('error.dagValidation'), variant: 'destructive' })
    }
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

async function onActivate() {
  if (!selectedVersionId.value) return
  try {
    await activateVersion(selectedVersionId.value)
    toast({ title: t('editor.activateVersion') })
    versions.value = await fetchVersions(workflowId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
}

async function onDeactivate() {
  if (!selectedVersionId.value) return
  try {
    await deactivateVersion(selectedVersionId.value)
    deactivateOpen.value = false
    toast({ title: t('editor.deactivateVersion') })
    versions.value = await fetchVersions(workflowId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
}

async function onCreateVersion() {
  try {
    const ver = await createVersion(workflowId.value)
    versions.value = await fetchVersions(workflowId.value)
    selectedVersionId.value = ver.id
    toast({ title: t('editor.createVersion') })
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
}

async function onNameBlur() {
  try {
    await updateWorkflow(workflowId.value, { name: workflowName.value })
  }
  catch {
    // Non-critical
  }
}

function onConnect(params: Connection) {
  dagEditor.connectSteps(params)
}

function onNodeDragStop(nodeId: string, position: { x: number; y: number }) {
  dagEditor.onNodeDragStop(nodeId, position)
}

function onNodeClick(nodeId: string) {
  dagEditor.selectNode(nodeId)
}

function onNodeDoubleClick(nodeId: string) {
  nodeEditor.open(nodeId)
}

function onEdgeClick(_edgeId: string) {
  // Select edge for potential deletion
}

function onRemoveEdge(edgeId: string) {
  dagEditor.removeEdge(edgeId)
}

function onDrop(
  stepType: string,
  workTypeId: number | undefined,
  workTypeCode: string | undefined,
  position: { x: number; y: number },
  name: string | undefined,
) {
  dagEditor.addStep(stepType, workTypeId, workTypeCode, position, name)
}

function onToolbarAddStep(
  stepType: string,
  workTypeId: number | undefined,
  position: { x: number; y: number },
) {
  dagEditor.addStep(stepType, workTypeId, undefined, position)
}

function onDeleteSelected() {
  if (dagEditor.selectedNodeId.value) {
    dagEditor.removeStep(dagEditor.selectedNodeId.value)
  }
}

function onPanelUpdateData(nodeId: string, data: Partial<StepData>) {
  dagEditor.updateNodeData(nodeId, data)
}

function onPanelUpdateStep(stepId: string, data: Record<string, unknown>) {
  dagEditor.updateStepOnServer(stepId, data)
}

function onPanelDeleteStep(stepId: string) {
  dagEditor.removeStep(stepId)
}

function onPanelOpenEditor(nodeId: string) {
  nodeEditor.open(nodeId)
}

function onNodeEditorSave(nodeId: string, config: Record<string, unknown>, inputMapping: Record<string, string>) {
  dagEditor.updateNodeData(nodeId, { config, inputMapping })
  dagEditor.updateStepOnServer(nodeId, { config, input_mapping: inputMapping })
}

function dismissErrors() {
  dagEditor.validationErrors.value = []
}

onMounted(() => {
  loadAll()
})
</script>

<template>
  <div class="flex h-[calc(100vh-3.5rem)] flex-col">
    <!-- Header bar -->
    <div class="flex items-center gap-3 border-b px-4 py-2">
      <Input
        v-model="workflowName"
        class="h-8 w-64 border-0 p-0 text-sm font-semibold shadow-none focus-visible:ring-0"
        @blur="onNameBlur"
        @keydown.enter="($event.target as HTMLInputElement)?.blur()"
      />

      <VersionSelector
        v-model="selectedVersionId"
        :versions="versions"
      />

      <Button
        variant="outline"
        size="sm"
        @click="onCreateVersion"
      >
        {{ t('editor.createVersion') }}
      </Button>

      <div class="flex-1" />

      <Button
        size="sm"
        :disabled="saving || !selectedVersionId"
        @click="onSave"
      >
        <Save class="mr-2 h-4 w-4" />
        {{ t('editor.saveVersion') }}
      </Button>

      <Button
        v-if="!isVersionActive"
        size="sm"
        variant="outline"
        :disabled="!selectedVersionId"
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

    <!-- Validation errors -->
    <div
      v-if="dagEditor.validationErrors.value.length > 0"
      class="mx-4 mt-2 rounded-md border border-red-300 bg-red-50 p-3 dark:border-red-800 dark:bg-red-950"
    >
      <div class="flex items-start justify-between">
        <div class="flex items-start gap-2">
          <AlertCircle class="mt-0.5 h-4 w-4 text-red-600 dark:text-red-400" />
          <div>
            <p class="text-sm font-medium text-red-800 dark:text-red-200">
              {{ t('editor.validationErrors') }}
            </p>
            <ul class="mt-1 list-disc pl-4 text-sm text-red-700 dark:text-red-300">
              <li
                v-for="(error, i) in dagEditor.validationErrors.value"
                :key="i"
              >
                {{ error }}
              </li>
            </ul>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 shrink-0"
          @click="dismissErrors"
        >
          <X class="h-4 w-4" />
        </Button>
      </div>
    </div>

    <!-- Editor layout -->
    <div v-if="!pageLoading && versions.length > 0" class="flex flex-1 overflow-hidden">
      <!-- Left toolbar -->
      <StepToolbar @add-step="onToolbarAddStep" />

      <!-- Center canvas -->
      <div class="flex-1">
        <DagCanvas
          mode="edit"
          :nodes="dagEditor.nodes.value"
          :edges="dagEditor.edges.value"
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
        v-if="dagEditor.selectedNode.value"
        :node="dagEditor.selectedNode.value"
        :work-types="workTypes"
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
        :cta-label="t('empty.versions.cta')"
        @cta="onCreateVersion"
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
      :all-nodes="dagEditor.nodes.value"
      :all-edges="dagEditor.edges.value"
      @save="onNodeEditorSave"
    />

    <!-- Deactivate Confirmation Dialog -->
    <Dialog v-model:open="deactivateOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('destructive.deactivateVersion.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deactivateVersion.body', { number: currentVersion?.version_number ?? '' }) }}
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

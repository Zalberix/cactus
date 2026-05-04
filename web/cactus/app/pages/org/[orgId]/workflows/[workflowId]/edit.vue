<script setup lang="ts">
import '@vue-flow/core/dist/style.css'
import { Save, Play, Pause, AlertCircle, X, Trash2, Key, Download } from 'lucide-vue-next'
import type { Connection } from '@vue-flow/core'
import type { Version, WorkType } from '~/composables/useVersions'
import type { StepData } from '~/composables/useDagEditor'
import type { Token, CreateTokenResponse } from '~/composables/useSystems'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import StepToolbar from '~/components/dag/StepToolbar.vue'
import StepPanel from '~/components/dag/StepPanel.vue'
import VersionSelector from '~/components/dag/VersionSelector.vue'
import NodeEditor from '~/components/dag/node-editor/NodeEditor.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import { Button } from '~/components/ui/button'
import { Checkbox } from '~/components/ui/checkbox'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
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
  fetchTokens,
  createToken,
  fetchWorkflowTokens,
  bindTokenWorkflow,
  unbindTokenWorkflow,
} = useSystems()
const {
  fetchVersions,
  createVersion,
  activateVersion,
  deactivateVersion,
  deleteVersion,
  fetchWorkTypes,
} = useVersions()

// Workflow state
const workflowName = ref('')
const workflowSystemId = ref<number | null>(null)
const versions = ref<Version[]>([])
const selectedVersionId = ref<number | null>(null)
const workTypes = ref<WorkType[]>([])
const pageLoading = ref(true)
const saving = ref(false)
const deactivateOpen = ref(false)
const deleteVersionOpen = ref(false)
const deletingVersion = ref(false)
const tokensOpen = ref(false)
const workflowTokens = ref<Token[]>([])
const originalTokenIds = ref<Set<number>>(new Set())
const selectedTokenIds = ref<Set<number>>(new Set())
const tokensLoading = ref(false)
const tokensSaving = ref(false)
const tokenCreateName = ref('')
const tokenCreating = ref(false)
const createdToken = ref<CreateTokenResponse | null>(null)

// DAG editor composable
const dagEditor = useDagEditor(workflowId, selectedVersionId)

// Node editor composable
const nodeEditor = useNodeEditor()

const canvasEdges = computed(() =>
  dagEditor.edges.value.map(edge => ({
    ...edge,
    selected: edge.id === dagEditor.selectedEdgeId.value,
  })),
)

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
    workflowSystemId.value = wf.system_id
    versions.value = vers
    workTypes.value = wts

    // Auto-select: prefer active version, else latest
    if (vers.length > 0) {
      const active = vers.find(v => v.is_active)
      const latest = vers[vers.length - 1]
      selectedVersionId.value = (active ?? latest).id
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
  if (vid) {
    try {
      await dagEditor.loadSteps(vid)
    }
    catch (err) {
      toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
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
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
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
    versions.value = await fetchVersions(workflowId.value)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onCreateVersion() {
  try {
    const ver = await createVersion(workflowId.value)
    versions.value = await fetchVersions(workflowId.value)
    selectedVersionId.value = ver.id
    toast({ title: t('editor.createVersion') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onDeleteVersion() {
  const versionId = selectedVersionId.value
  if (!versionId) return

  deletingVersion.value = true
  try {
    await deleteVersion(versionId)
    deleteVersionOpen.value = false

    const nextVersions = await fetchVersions(workflowId.value)
    versions.value = nextVersions

    if (nextVersions.length > 0) {
      const active = nextVersions.find(v => v.is_active)
      const latest = nextVersions[nextVersions.length - 1]
      selectedVersionId.value = (active ?? latest).id
    }
    else {
      selectedVersionId.value = null
      dagEditor.nodes.value = []
      dagEditor.edges.value = []
      dagEditor.selectNode(null)
      dagEditor.selectEdge(null)
    }

    toast({ title: t('editor.deleteVersion') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    deletingVersion.value = false
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
  const node = dagEditor.nodes.value.find(n => n.id === nodeId)
  if (node?.data.controlKind === 'start') return
  nodeEditor.open(nodeId)
}

function onEdgeClick(edgeId: string) {
  dagEditor.selectEdge(edgeId)
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
  workTypeCode: string | undefined,
  position: { x: number; y: number },
  name: string | undefined,
) {
  dagEditor.addStep(stepType, workTypeId, workTypeCode, position, name)
}

function onDeleteSelected() {
  if (dagEditor.selectedEdgeId.value) {
    dagEditor.removeEdge(dagEditor.selectedEdgeId.value)
    return
  }

  if (dagEditor.selectedNodeId.value) {
    dagEditor.removeStep(dagEditor.selectedNodeId.value)
  }
}

async function openTokensDialog() {
  if (!workflowSystemId.value) return
  tokensOpen.value = true
  tokensLoading.value = true
  createdToken.value = null
  tokenCreateName.value = ''
  try {
    const [tokens, links] = await Promise.all([
      fetchTokens(workflowSystemId.value),
      fetchWorkflowTokens(workflowId.value),
    ])
    workflowTokens.value = tokens
    const linkedIds = new Set(links.map(link => link.system_token_id))
    originalTokenIds.value = new Set(linkedIds)
    selectedTokenIds.value = new Set(linkedIds)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    tokensLoading.value = false
  }
}

function isTokenSelected(tokenId: number) {
  return selectedTokenIds.value.has(tokenId)
}

function toggleToken(tokenId: number) {
  const next = new Set(selectedTokenIds.value)
  if (next.has(tokenId)) {
    next.delete(tokenId)
  }
  else {
    next.add(tokenId)
  }
  selectedTokenIds.value = next
}

async function saveWorkflowTokenAccess() {
  tokensSaving.value = true
  try {
    const selected = selectedTokenIds.value
    const original = originalTokenIds.value
    const toBind = [...selected].filter(id => !original.has(id))
    const toUnbind = [...original].filter(id => !selected.has(id))

    await Promise.all([
      ...toBind.map(tokenId => bindTokenWorkflow(tokenId, workflowId.value)),
      ...toUnbind.map(tokenId => unbindTokenWorkflow(tokenId, workflowId.value)),
    ])

    tokensOpen.value = false
    toast({ title: t('tokens.workflowAccessSaved') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    tokensSaving.value = false
  }
}

async function createTokenForWorkflow() {
  if (!workflowSystemId.value || !tokenCreateName.value.trim()) return
  tokenCreating.value = true
  try {
    const token = await createToken(workflowSystemId.value, tokenCreateName.value.trim())
    await bindTokenWorkflow(token.id, workflowId.value)
    createdToken.value = token
    workflowTokens.value = await fetchTokens(workflowSystemId.value)
    originalTokenIds.value = new Set([...originalTokenIds.value, token.id])
    selectedTokenIds.value = new Set([...selectedTokenIds.value, token.id])
    tokenCreateName.value = ''
    toast({ title: t('tokens.created') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    tokenCreating.value = false
  }
}

function downloadTokenJson() {
  if (!createdToken.value) return
  const data = {
    public_token: createdToken.value.public_token,
    private_token: createdToken.value.private_token,
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `token-${createdToken.value.public_token.slice(0, 8)}.json`
  a.click()
  URL.revokeObjectURL(url)
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
  const node = dagEditor.nodes.value.find(n => n.id === nodeId)
  if (node?.data.controlKind === 'start') return
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

      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8 text-destructive hover:text-destructive"
        :disabled="!selectedVersionId || deletingVersion"
        :title="t('editor.deleteVersion')"
        @click="deleteVersionOpen = true"
      >
        <Trash2 class="h-4 w-4" />
      </Button>

      <div class="flex-1" />

      <Button
        size="sm"
        variant="outline"
        class="gap-2"
        :disabled="!workflowSystemId"
        @click="openTokensDialog"
      >
        <Key class="h-4 w-4" />
        {{ t('tokens.title') }}
      </Button>

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

    <!-- Workflow Tokens Dialog -->
    <Dialog v-model:open="tokensOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('tokens.workflowDialogTitle') }}</DialogTitle>
          <DialogDescription>{{ t('tokens.workflowDialogDescription') }}</DialogDescription>
        </DialogHeader>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label>{{ t('tokens.createForWorkflow') }}</Label>
            <div class="flex gap-2">
              <Input
                v-model="tokenCreateName"
                :placeholder="t('tokens.namePlaceholder')"
                :disabled="tokenCreating"
              />
              <Button
                :disabled="tokenCreating || !tokenCreateName.trim()"
                @click="createTokenForWorkflow"
              >
                {{ t('common.create') }}
              </Button>
            </div>
          </div>

          <div
            v-if="createdToken"
            class="space-y-3 rounded-md border border-yellow-500/50 bg-yellow-500/10 p-3"
          >
            <p class="text-sm font-medium">{{ t('tokenShowOnce.heading') }}</p>
            <code class="block rounded-md border bg-background px-3 py-2 text-xs font-mono break-all">
              {{ createdToken.public_token }}
            </code>
            <code class="block rounded-md border bg-background px-3 py-2 text-xs font-mono break-all">
              {{ createdToken.private_token }}
            </code>
            <Button variant="outline" size="sm" class="gap-2" @click="downloadTokenJson">
              <Download class="h-4 w-4" />
              {{ t('tokens.downloadJson') }}
            </Button>
          </div>

          <div class="max-h-[320px] space-y-2 overflow-y-auto pr-1">
            <p v-if="tokensLoading" class="text-sm text-muted-foreground">
              {{ t('common.loading') }}
            </p>
            <p
              v-else-if="workflowTokens.length === 0"
              class="text-sm text-muted-foreground"
            >
              {{ t('tokens.noTokens') }}
            </p>
            <template v-else>
              <label
                v-for="token in workflowTokens"
                :key="token.id"
                class="flex cursor-pointer items-center gap-3 rounded-md border px-3 py-2 text-sm hover:bg-muted/50"
              >
                <Checkbox
                  :checked="isTokenSelected(token.id)"
                  @update:checked="toggleToken(token.id)"
                />
                <span class="min-w-0 flex-1">
                  <span class="block font-medium">{{ token.name || token.public_token }}</span>
                  <span class="block truncate text-xs text-muted-foreground">{{ token.public_token }}</span>
                </span>
                <span
                  class="text-xs"
                  :class="token.is_active ? 'text-green-600' : 'text-muted-foreground'"
                >
                  {{ token.is_active ? t('status.active') : t('status.inactive') }}
                </span>
              </label>
            </template>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="tokensOpen = false">
            {{ t('common.cancel') }}
          </Button>
          <Button
            :disabled="tokensLoading || tokensSaving"
            @click="saveWorkflowTokenAccess"
          >
            {{ t('common.save') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

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

    <!-- Delete Version Confirmation Dialog -->
    <Dialog v-model:open="deleteVersionOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('destructive.deleteVersion.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deleteVersion.body', { number: currentVersion?.version_number ?? '' }) }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" @click="deleteVersionOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button
            variant="destructive"
            :disabled="deletingVersion"
            @click="onDeleteVersion"
          >
            {{ t('destructive.deleteVersion.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { GitBranch, Key, Route } from 'lucide-vue-next'
import type { VersionSummary } from '~/composables/useVersions'
import type { WorkflowInputSchemaRecord, WorkflowSchemaCompatibility } from '~/composables/useWorkflowRouting'
import { buildSupportedSchemaLabelsByVersionId } from '~/components/dag/supported-schema-utils'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'
import { nativeInputSchemaForVersion, workflowInputSchemaEditorPath } from '~/composables/useWorkflowRouting'
import WorkflowVersionCreateMenu from '~/components/dag/WorkflowVersionCreateMenu.vue'
import WorkflowVersionTable from '~/components/dag/WorkflowVersionTable.vue'
import WorkflowTokensDialog from '~/components/dag/WorkflowTokensDialog.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const { fetchWorkflow, updateWorkflow } = useWorkflows()
const {
  fetchVersionSummaries,
  createVersion,
  activateVersion,
  deactivateVersion,
  deleteVersion,
} = useVersions()
const {
  fetchInputSchemas,
  fetchCompatibilities,
} = useWorkflowRouting()

const workflowName = ref('')
const workflowSystemId = ref<number | null>(null)
const workflowPriority = ref(1)
const versions = ref<VersionSummary[]>([])
const inputSchemas = ref<WorkflowInputSchemaRecord[]>([])
const compatibilities = ref<WorkflowSchemaCompatibility[]>([])
const loading = ref(true)
const tokensOpen = ref(false)
const supportedSchemaLabelsByVersionId = computed(() =>
  buildSupportedSchemaLabelsByVersionId(inputSchemas.value, compatibilities.value),
)

async function loadVersionRoutingData() {
  const [summaries, schemas, schemaCompatibilities] = await Promise.all([
    fetchVersionSummaries(workflowId.value),
    fetchInputSchemas(workflowId.value),
    fetchCompatibilities(workflowId.value),
  ])
  versions.value = summaries
  inputSchemas.value = schemas
  compatibilities.value = schemaCompatibilities
}

async function loadData() {
  loading.value = true
  try {
    const [workflow] = await Promise.all([
      fetchWorkflow(workflowId.value),
      loadVersionRoutingData(),
    ])
    workflowName.value = workflow.name
    workflowSystemId.value = workflow.system_id
    workflowPriority.value = workflow.priority
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function onNameBlur() {
  try {
    await updateWorkflow(workflowId.value, { name: workflowName.value, priority: workflowPriority.value })
  }
  catch {
    // Best-effort inline rename.
  }
}

async function onActivate(version: VersionSummary) {
  try {
    await activateVersion(version.id)
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onDeactivate(version: VersionSummary) {
  try {
    await deactivateVersion(version.id)
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onDelete(version: VersionSummary) {
  try {
    await deleteVersion(version.id)
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onCreateInitialVersion() {
  try {
    const version = await createVersion(workflowId.value)
    await router.push(workflowVersionEditorPath(orgId.value, workflowId.value, version.id))
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

async function onShowInputSchema(version: VersionSummary) {
  try {
    if (inputSchemas.value.length === 0 || compatibilities.value.length === 0) {
      await loadVersionRoutingData()
    }
    const schema = nativeInputSchemaForVersion(version.id, inputSchemas.value, compatibilities.value)
    if (!schema) {
      throw new Error(t('workflowRouting.errorNativeSchemaMissing'))
    }
    await router.push(workflowInputSchemaEditorPath(orgId.value, workflowId.value, schema.id))
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 border-b pb-4 md:flex-row md:items-center">
      <div class="min-w-0 flex-1">
        <Input
          v-model="workflowName"
          class="h-10 max-w-xl border-0 px-0 text-2xl font-semibold shadow-none focus-visible:ring-0"
          @blur="onNameBlur"
          @keydown.enter="($event.target as HTMLInputElement)?.blur()"
        />
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <WorkflowVersionCreateMenu
          :workflow-id="workflowId"
          :org-id="orgId"
          :versions="versions"
          @created="loadData"
        />
        <Button variant="outline" class="gap-2" @click="router.push(`/org/${orgId}/workflows/${workflowId}/routing`)">
          <GitBranch class="h-4 w-4" />
          {{ t('workflowRouting.nav') }}
        </Button>
        <Button variant="outline" class="gap-2" :disabled="!workflowSystemId" @click="tokensOpen = true">
          <Key class="h-4 w-4" />
          {{ t('workflowVersions.tokens') }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>

    <WorkflowVersionTable
      v-else-if="versions.length > 0"
      :versions="versions"
      :org-id="orgId"
      :workflow-id="workflowId"
      :supported-schema-labels-by-version-id="supportedSchemaLabelsByVersionId"
      @activate="onActivate"
      @deactivate="onDeactivate"
      @delete="onDelete"
      @show-input-schema="onShowInputSchema"
    />

    <EmptyState
      v-else
      :icon="Route"
      :heading="t('empty.versions.heading')"
      :body="t('empty.versions.body')"
      :cta-label="t('empty.versions.cta')"
      @cta="onCreateInitialVersion"
    />

    <WorkflowTokensDialog
      v-model:open="tokensOpen"
      :workflow-id="workflowId"
      :system-id="workflowSystemId"
    />
  </div>
</template>

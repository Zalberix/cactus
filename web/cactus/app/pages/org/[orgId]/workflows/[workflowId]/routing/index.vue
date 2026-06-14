<script setup lang="ts">
import type { RoutingVersionRowRecord } from '~/composables/useWorkflowRouting'
import { ArrowLeft } from 'lucide-vue-next'
import {
  workflowInputSchemaCompatibilitiesPath,
  workflowInputSchemaCreatePath,
  workflowInputSchemaEditorPath,
  workflowRoutingTestingPath,
} from '~/composables/useWorkflowRouting'
import RoutingArchiveSchemaDialog from '~/components/dag/routing/RoutingArchiveSchemaDialog.vue'
import RoutingVersionTable from '~/components/dag/routing/RoutingVersionTable.vue'
import { Button } from '~/components/ui/button'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const { fetchWorkflow } = useWorkflows()
const routing = useWorkflowRouting()

const workflowName = ref('')
const rows = ref<RoutingVersionRowRecord[]>([])
const loading = ref(true)
const saving = ref(false)
const page = ref(1)
const pageCount = ref(1)
const pageSize = ref(20)
const archiveOpen = ref(false)
const archiveRow = ref<RoutingVersionRowRecord | null>(null)

async function loadData() {
  loading.value = true
  try {
    const [workflow, routingRows] = await Promise.all([
      fetchWorkflow(workflowId.value),
      routing.fetchRoutingVersionRows(workflowId.value, page.value, pageSize.value),
    ])
    workflowName.value = workflow.name
    rows.value = routingRows.data
    pageCount.value = routingRows.meta.total_pages
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function editRow(row: RoutingVersionRowRecord) {
  router.push(workflowInputSchemaEditorPath(orgId.value, workflowId.value, row.input_schema_id))
}

function openCompatibility(row: RoutingVersionRowRecord) {
  router.push(workflowInputSchemaCompatibilitiesPath(orgId.value, workflowId.value, row.input_schema_id))
}

function openArchive(row: RoutingVersionRowRecord) {
  archiveRow.value = row
  archiveOpen.value = true
}

function inputSchemaEditorPath(row: RoutingVersionRowRecord) {
  return workflowInputSchemaEditorPath(orgId.value, workflowId.value, row.input_schema_id)
}

async function confirmArchive(requiredLabel: string) {
  if (!archiveRow.value) return
  saving.value = true
  try {
    await routing.archiveInputSchema(archiveRow.value.input_schema_id, requiredLabel)
    toast({ title: t('workflowRouting.archivedInputSchema') })
    archiveOpen.value = false
    archiveRow.value = null
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('workflowRouting.errorArchiveSchema')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadData()
})

watch([page, pageSize], () => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
      <div>
        <Button variant="ghost" class="-ml-3 mb-3 gap-2" @click="router.push(`/org/${orgId}/workflows/${workflowId}`)">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <div>
          <h1 class="text-2xl font-semibold">{{ t('workflowRouting.routingVersions') }}</h1>
          <p class="text-sm text-muted-foreground">
            {{ workflowName || t('common.loading') }}
          </p>
        </div>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button type="button" @click="router.push(workflowInputSchemaCreatePath(orgId, workflowId))">
          {{ t('workflowRouting.createInputSchema') }}
        </Button>
        <Button type="button" variant="outline" @click="router.push(workflowRoutingTestingPath(orgId, workflowId))">
          {{ t('workflowRouting.testing') }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>
    <RoutingVersionTable
      v-else
      :rows="rows"
      :saving="saving"
      :edit-path="inputSchemaEditorPath"
      v-model:page="page"
      v-model:page-size="pageSize"
      :page-count="pageCount"
      @edit="editRow"
      @compatibility="openCompatibility"
      @archive="openArchive"
    />

    <RoutingArchiveSchemaDialog
      v-model:open="archiveOpen"
      :row="archiveRow"
      :saving="saving"
      @confirm="confirmArchive"
    />
  </div>
</template>

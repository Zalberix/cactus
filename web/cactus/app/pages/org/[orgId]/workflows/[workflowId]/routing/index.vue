<script setup lang="ts">
import type { RoutingVersionRowRecord } from '~/composables/useWorkflowRouting'
import { ArrowLeft, Route } from 'lucide-vue-next'
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
const archiveOpen = ref(false)
const archiveRow = ref<RoutingVersionRowRecord | null>(null)

async function loadData() {
  loading.value = true
  try {
    const [workflow, routingRows] = await Promise.all([
      fetchWorkflow(workflowId.value),
      routing.fetchRoutingVersionRows(workflowId.value),
    ])
    workflowName.value = workflow.name
    rows.value = routingRows
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function editRow(row: RoutingVersionRowRecord) {
  router.push(workflowInputSchemaEditorPath(orgId.value, workflowId.value, row.native_input_schema_id))
}

function openCompatibility(row: RoutingVersionRowRecord) {
  router.push(workflowInputSchemaCompatibilitiesPath(orgId.value, workflowId.value, row.native_input_schema_id))
}

function openArchive(row: RoutingVersionRowRecord) {
  archiveRow.value = row
  archiveOpen.value = true
}

async function confirmArchive(requiredLabel: string) {
  if (!archiveRow.value) return
  saving.value = true
  try {
    await routing.archiveInputSchema(archiveRow.value.native_input_schema_id, requiredLabel)
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
              <h1 class="text-2xl font-semibold">{{ t('workflowRouting.routingVersions') }}</h1>
              <p class="text-sm text-muted-foreground">
                {{ workflowName || t('common.loading') }}
              </p>
            </div>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          <Button type="button" variant="outline" @click="router.push(workflowRoutingTestingPath(orgId, workflowId))">
            {{ t('workflowRouting.testing') }}
          </Button>
          <Button type="button" @click="router.push(workflowInputSchemaCreatePath(orgId, workflowId))">
            {{ t('workflowRouting.createInputSchema') }}
          </Button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>
    <RoutingVersionTable
      v-else
      :rows="rows"
      :saving="saving"
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

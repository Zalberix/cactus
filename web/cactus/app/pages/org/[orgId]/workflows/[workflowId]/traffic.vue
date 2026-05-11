<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import { ArrowLeft } from 'lucide-vue-next'
import WorkflowTrafficSettings from '~/components/dag/WorkflowTrafficSettings.vue'
import { Button } from '~/components/ui/button'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))

const { fetchWorkflow } = useWorkflows()
const { fetchVersionSummaries, updateWorkflowTraffic } = useVersions()

const workflowName = ref('')
const versions = ref<VersionSummary[]>([])
const loading = ref(true)
const saving = ref(false)

const activeVersions = computed(() =>
  versions.value.filter(version => version.is_active && !version.deleted_at),
)

async function loadData() {
  loading.value = true
  try {
    const [workflow, summaries] = await Promise.all([
      fetchWorkflow(workflowId.value),
      fetchVersionSummaries(workflowId.value),
    ])
    workflowName.value = workflow.name
    versions.value = summaries
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function onSave(payload: { versions: Array<{ version_id: number, mode: 'share' | 'fixed', weight?: number }> }) {
  saving.value = true
  try {
    await updateWorkflowTraffic(workflowId.value, payload)
    toast({ title: t('workflowTraffic.saved') })
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 border-b pb-4 md:flex-row md:items-center md:justify-between">
      <div>
        <Button variant="ghost" class="-ml-3 mb-2 gap-2" @click="router.push(`/org/${orgId}/workflows/${workflowId}`)">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <h1 class="text-2xl font-semibold">{{ t('workflowTraffic.title') }}</h1>
        <p class="text-sm text-muted-foreground">
          {{ workflowName || t('common.loading') }}
        </p>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('common.loading') }}
    </div>

    <WorkflowTrafficSettings
      v-else
      :versions="activeVersions"
      :saving="saving"
      @save="onSave"
    />
  </div>
</template>

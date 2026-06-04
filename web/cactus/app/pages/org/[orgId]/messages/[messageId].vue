<script setup lang="ts">
import { ArrowLeft } from 'lucide-vue-next'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import WsIndicator from '~/components/feedback/WsIndicator.vue'
import MessageStepDetailsPanel from '~/components/messages/MessageStepDetailsPanel.vue'
import { formatMessageWorkflowSubtitle } from '~/components/messages/message-version-utils'
import { Button } from '~/components/ui/button'
import { Skeleton } from '~/components/ui/skeleton'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const messageId = computed(() => Number(route.params.messageId))

const {
  detail,
  runSteps,
  nodes,
  edges,
  workflowStatus,
  isConnected,
  error,
  selectedStep,
  initialLoading,
} = useDagViewer(messageId)

watch(
  () => detail.value?.graph.steps,
  (steps) => {
    if (!selectedStep.value) {
      const firstStep = steps?.[0]
      if (firstStep) selectedStep.value = firstStep.id
    }
  },
  { immediate: true },
)

const selectedGraphStep = computed(() => {
  const step = detail.value?.graph.steps.find(item => item.id === selectedStep.value)
  return step ? { ...step, input_mapping: [...(step.input_mapping ?? [])] } : null
})

const selectedRunStep = computed(() => {
  const step = selectedStep.value ? runSteps.value.get(selectedStep.value) : undefined
  return step ? { ...step } : undefined
})
const runStepsForPanel = computed(() => new Map(runSteps.value))

const failedStep = computed(() =>
  Array.from(runSteps.value.values()).find(step => step.status === 'failed'),
)

const workflowFailureText = computed(() => {
  if (workflowStatus.value !== 'failed') return ''
  if (failedStep.value) {
    return t('messages.detail.failedAtStep', { id: failedStep.value.step_id })
  }
  return detail.value?.workflow_run?.error_message ?? error.value ?? ''
})

const workflowSubtitle = computed(() =>
  detail.value ? formatMessageWorkflowSubtitle(detail.value) : '',
)

const versionInputData = computed(() =>
  detail.value?.version_input_data ?? detail.value?.workflow_run?.version_input_data,
)

const routingDecision = computed(() =>
  detail.value?.routing_decision ?? detail.value?.workflow_run?.routing_decision,
)

const schemaLabel = computed(() => {
  const code = detail.value?.workflow_input_schema_code?.trim()
  const version = typeof detail.value?.workflow_input_schema_version === 'number'
    ? `v${detail.value.workflow_input_schema_version}`
    : ''

  if (code && version) return `${code} ${version}`
  if (code) return code
  if (version) return version
  return detail.value?.workflow_input_schema_id ? `#${detail.value.workflow_input_schema_id}` : '-'
})

const routingMeta = computed(() => {
  const current = detail.value
  if (!current) return []
  return [
    { label: t('messages.detail.schema'), value: schemaLabel.value },
    { label: t('messages.detail.selectionReason'), value: current.selection_reason ?? current.workflow_run?.selection_reason ?? '-' },
    { label: t('messages.detail.compatibility'), value: current.input_schema_compatibility_id ?? current.workflow_run?.input_schema_compatibility_id ?? '-' },
    { label: t('messages.detail.experiment'), value: current.experiment_id ?? current.workflow_run?.experiment_id ?? '-' },
    { label: t('messages.detail.scope'), value: current.experiment_scope_id ?? current.workflow_run?.experiment_scope_id ?? '-' },
    { label: t('messages.detail.variant'), value: current.experiment_variant_id ?? current.workflow_run?.experiment_variant_id ?? '-' },
  ]
})

function formatJson(value: unknown) {
  if (value === undefined || value === null) return '-'
  return JSON.stringify(value, null, 2)
}

function mapStatus(status: string): 'pending' | 'running' | 'done' | 'error' | 'working' {
  if (status === 'completed' || status === 'done') return 'done'
  if (status === 'failed' || status === 'error') return 'error'
  if (status === 'skipped') return 'working'
  if (status === 'running') return 'running'
  return 'pending'
}

function onNodeClick(nodeId: string) {
  selectedStep.value = Number(nodeId)
}

function goBack() {
  router.push(`/org/${orgId.value}/messages`)
}
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex min-w-0 items-center gap-3">
        <Button variant="ghost" size="icon" @click="goBack">
          <ArrowLeft class="size-4" />
        </Button>
        <div class="min-w-0">
          <h1 class="truncate text-lg font-semibold">
            {{ t('messages.detail.title', { id: messageId }) }}
          </h1>
          <p v-if="workflowSubtitle" class="truncate text-xs text-muted-foreground">
            {{ workflowSubtitle }}
          </p>
        </div>
        <StatusBadge :status="mapStatus(workflowStatus)" />
      </div>

      <WsIndicator :connected="isConnected" />
    </div>

    <div
      v-if="workflowFailureText"
      class="shrink-0 border-b border-red-200 bg-red-50 px-4 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200"
    >
      {{ workflowFailureText }}
    </div>

    <div
      v-if="detail"
      class="grid shrink-0 gap-3 border-b bg-muted/20 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_320px]"
    >
      <section class="min-w-0 rounded-md border bg-background p-3">
        <h2 class="mb-2 text-sm font-semibold">{{ t('messages.detail.publicInput') }}</h2>
        <pre class="max-h-44 overflow-auto whitespace-pre-wrap break-words rounded bg-muted/40 p-3 text-xs">{{ formatJson(detail.message_value) }}</pre>
      </section>

      <section class="min-w-0 rounded-md border bg-background p-3">
        <h2 class="mb-2 text-sm font-semibold">{{ t('messages.detail.mappedInput') }}</h2>
        <pre class="max-h-44 overflow-auto whitespace-pre-wrap break-words rounded bg-muted/40 p-3 text-xs">{{ formatJson(versionInputData) }}</pre>
      </section>

      <section class="rounded-md border bg-background p-3">
        <h2 class="mb-2 text-sm font-semibold">{{ t('messages.detail.routing') }}</h2>
        <dl class="grid grid-cols-[120px_minmax(0,1fr)] gap-x-3 gap-y-2 text-xs">
          <template v-for="item in routingMeta" :key="item.label">
            <dt class="text-muted-foreground">{{ item.label }}</dt>
            <dd class="truncate font-mono">{{ item.value }}</dd>
          </template>
        </dl>
        <pre
          v-if="routingDecision"
          class="mt-3 max-h-28 overflow-auto whitespace-pre-wrap break-words rounded bg-muted/40 p-3 text-xs"
        >{{ formatJson(routingDecision) }}</pre>
      </section>
    </div>

    <div v-if="initialLoading" class="flex flex-1 items-center justify-center">
      <div class="w-full max-w-lg space-y-4">
        <Skeleton class="h-48 w-full" />
        <Skeleton class="h-8 w-full" />
        <Skeleton class="h-8 w-full" />
      </div>
    </div>

    <div v-else class="grid min-h-0 flex-1 grid-cols-[minmax(0,1fr)_420px]">
      <div class="min-h-0">
        <DagCanvas
          mode="view"
          show-runtime-state
          :nodes="nodes"
          :edges="edges"
          @node-click="onNodeClick"
        />
      </div>
      <MessageStepDetailsPanel
        :step="selectedGraphStep"
        :runtime="selectedRunStep"
        :message-value="detail?.message_value ?? {}"
        :run-steps="runStepsForPanel"
      />
    </div>

    <div
      v-if="error && !workflowFailureText"
      class="fixed bottom-4 right-4 max-w-sm rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950 dark:text-red-200"
    >
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowLeft } from 'lucide-vue-next'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import WsIndicator from '~/components/feedback/WsIndicator.vue'
import MessageStepDetailsPanel from '~/components/messages/MessageStepDetailsPanel.vue'
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
    if (!selectedStep.value && steps?.length) {
      selectedStep.value = steps[0].id
    }
  },
  { immediate: true },
)

const selectedGraphStep = computed(() =>
  detail.value?.graph.steps.find(step => step.id === selectedStep.value) ?? null,
)

const selectedRunStep = computed(() =>
  selectedStep.value ? runSteps.value.get(selectedStep.value) : undefined,
)

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
          <p v-if="detail?.workflow_name" class="truncate text-xs text-muted-foreground">
            {{ detail.workflow_name }}
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
        :run-steps="runSteps"
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

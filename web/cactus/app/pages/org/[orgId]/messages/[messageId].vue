<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import type { WsStepStatus } from '~/composables/useWebSocketStatus'
import DagCanvas from '~/components/dag/DagCanvas.vue'
import WsIndicator from '~/components/feedback/WsIndicator.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import { Button } from '~/components/ui/button'
import { Skeleton } from '~/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '~/components/ui/tooltip'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const orgId = computed(() => Number(route.params.orgId))
const messageId = computed(() => Number(route.params.messageId))

const {
  nodes,
  edges,
  workflowStatus,
  isConnected,
  error,
  selectedStep,
  initialLoading,
} = useDagViewer(messageId)

// Convert Map-based steps from WS to flat list for the table
const stepsList = computed(() => {
  const list: Array<WsStepStatus & { duration?: string }> = []
  for (const node of nodes.value) {
    const data = node.data as Record<string, unknown>
    list.push({
      step_id: Number(node.id),
      step_type: (data.stepType as string) ?? '',
      status: (data.status as string) ?? 'pending',
      started_at: data.started_at as string | undefined,
      completed_at: data.completed_at as string | undefined,
      error: data.error_message as string | undefined,
      duration: computeDuration(
        data.started_at as string | undefined,
        data.completed_at as string | undefined,
      ),
    })
  }
  return list
})

function computeDuration(startedAt?: string, completedAt?: string): string {
  if (!startedAt) return '-'
  const start = new Date(startedAt).getTime()
  const end = completedAt ? new Date(completedAt).getTime() : Date.now()
  const diff = Math.max(0, end - start)

  if (diff < 1000) return `${diff}ms`
  if (diff < 60000) return `${(diff / 1000).toFixed(1)}s`
  return `${Math.floor(diff / 60000)}m ${Math.floor((diff % 60000) / 1000)}s`
}

function mapStatus(status: string): string {
  if (status === 'completed' || status === 'done') return 'done'
  if (status === 'failed') return 'error'
  return status
}

function onNodeClick(nodeId: string) {
  selectedStep.value = Number(nodeId)
}

function goBack() {
  router.push(`/org/${orgId.value}/messages`)
}

type StepRow = WsStepStatus & { duration?: string }

const columns: ColumnDef<StepRow>[] = [
  {
    accessorKey: 'step_id',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: 'ID' }),
    cell: ({ row }) => h('span', { class: 'font-mono text-sm' }, `#${row.getValue('step_id')}`),
    size: 60,
  },
  {
    accessorKey: 'step_type',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('editor.stepType') }),
    cell: ({ row }) => h('span', { class: 'text-sm' }, row.getValue('step_type')),
    size: 100,
  },
  {
    accessorKey: 'status',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.status') }),
    cell: ({ row }) => {
      const status = row.getValue('status') as string
      return h(StatusBadge, { status: mapStatus(status) as any })
    },
    size: 100,
  },
  {
    accessorKey: 'started_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.detail.startedAt') }),
    cell: ({ row }) => {
      const val = row.getValue('started_at') as string | undefined
      return h('span', { class: 'text-xs text-muted-foreground' }, val ? new Date(val).toLocaleTimeString() : '-')
    },
    size: 100,
  },
  {
    accessorKey: 'completed_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.detail.finishedAt') }),
    cell: ({ row }) => {
      const val = row.getValue('completed_at') as string | undefined
      return h('span', { class: 'text-xs text-muted-foreground' }, val ? new Date(val).toLocaleTimeString() : '-')
    },
    size: 100,
  },
  {
    accessorKey: 'duration',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.detail.duration') }),
    cell: ({ row }) => h('span', { class: 'text-xs font-mono' }, row.getValue('duration') || '-'),
    size: 80,
  },
  {
    accessorKey: 'error',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.detail.error') }),
    cell: ({ row }) => {
      const err = row.getValue('error') as string | undefined
      if (!err) return h('span', { class: 'text-xs text-muted-foreground' }, '-')
      return h(TooltipProvider, {}, () =>
        h(Tooltip, {}, {
          default: () => [
            h(TooltipTrigger, { asChild: true }, () =>
              h('span', {
                class: 'text-xs text-red-600 dark:text-red-400 truncate max-w-[200px] inline-block cursor-help',
              }, err),
            ),
            h(TooltipContent, { class: 'max-w-sm' }, () =>
              h('p', { class: 'text-xs break-words' }, err),
            ),
          ],
        }),
      )
    },
    size: 200,
  },
]
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between border-b px-4 py-3 shrink-0">
      <div class="flex items-center gap-3">
        <Button variant="ghost" size="icon" @click="goBack">
          <ArrowLeft class="size-4" />
        </Button>
        <h1 class="text-lg font-semibold">
          {{ t('messages.detail.title', { id: messageId }) }}
        </h1>
        <StatusBadge :status="mapStatus(workflowStatus) as any" />
      </div>

      <WsIndicator :connected="isConnected" />
    </div>

    <!-- Loading state -->
    <div v-if="initialLoading" class="flex-1 flex items-center justify-center">
      <div class="space-y-4 w-full max-w-lg">
        <Skeleton class="h-48 w-full" />
        <Skeleton class="h-8 w-full" />
        <Skeleton class="h-8 w-full" />
      </div>
    </div>

    <template v-else>
      <!-- DAG Canvas (top 60%) -->
      <div class="h-[60%] border-b relative">
        <DagCanvas
          mode="view"
          :nodes="nodes"
          :edges="edges"
          @node-click="onNodeClick"
        />
      </div>

      <!-- Steps Detail Table (bottom 40%) -->
      <div class="h-[40%] overflow-auto p-4">
        <h2 class="text-sm font-semibold mb-3">
          {{ t('messages.detail.steps') }}
        </h2>

        <div v-if="stepsList.length === 0" class="text-sm text-muted-foreground text-center py-8">
          {{ t('messages.detail.noSteps') }}
        </div>

        <DataTable
          v-else
          :columns="columns"
          :data="stepsList"
          :page-size="50"
        />
      </div>
    </template>

    <!-- Error display -->
    <div
      v-if="error"
      class="fixed bottom-4 right-4 max-w-sm rounded-md border border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950 px-4 py-3 text-sm text-red-800 dark:text-red-200"
    >
      {{ error }}
    </div>
  </div>
</template>

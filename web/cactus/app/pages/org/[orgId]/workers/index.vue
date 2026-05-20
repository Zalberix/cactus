<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Server, Eye, MoreHorizontal, Trash2 } from 'lucide-vue-next'
import type { Worker, WorkerWorkflowUsage } from '~/composables/useWorkers'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import { Button } from '~/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '~/components/ui/sheet'
import { Skeleton } from '~/components/ui/skeleton'
import { Badge } from '~/components/ui/badge'
import { toast } from '~/components/ui/toast/use-toast'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'

const { t, locale } = useI18n()
const route = useRoute()
const orgId = computed(() => Number(route.params.orgId))

const { fetchWorkersForOrg, fetchWorkerWorkflowUsages, deleteWorker } = useWorkers()

const workers = ref<Worker[]>([])
const loading = ref(true)
const deletingWorkerId = ref<number | null>(null)
const blockedUsages = ref<Record<number, WorkerWorkflowUsage[]>>({})

const sheetOpen = ref(false)
const selectedWorker = ref<Worker | null>(null)
const workerUsages = ref<Record<number, WorkerWorkflowUsage[]>>({})
const usageLoading = ref(false)

const selectedWorkerUsages = computed(() => {
  if (!selectedWorker.value) return []
  return workerUsages.value[selectedWorker.value.id] ?? []
})

async function loadWorkers() {
  loading.value = true
  try {
    workers.value = await fetchWorkersForOrg(orgId.value)
  }
  catch (err) {
    toast({
      title: getErrorMessage(err, t('error.server')),
      variant: 'destructive',
    })
  }
  finally {
    loading.value = false
  }
}

async function openWorkflowUsages(worker: Worker) {
  selectedWorker.value = worker
  sheetOpen.value = false
  sheetOpen.value = true

  usageLoading.value = true
  try {
    workerUsages.value = {
      ...workerUsages.value,
      [worker.id]: await fetchWorkerWorkflowUsages(worker.id),
    }
  }
  catch {
    workerUsages.value = {
      ...workerUsages.value,
      [worker.id]: [],
    }
  }
  finally {
    usageLoading.value = false
  }
}

async function onDeleteWorker(worker: Worker) {
  deletingWorkerId.value = worker.id
  try {
    const result = await deleteWorker(worker.id)
    if (result.result === 'deleted') {
      toast({ title: t('workers.deleted') })
      blockedUsages.value = { ...blockedUsages.value, [worker.id]: [] }
      await loadWorkers()
      return
    }

    if (result.result === 'in_use') {
      blockedUsages.value = {
        ...blockedUsages.value,
        [worker.id]: result.usages ?? [],
      }
      toast({ title: t('workers.inUse'), variant: 'destructive' })
      return
    }

    if (result.result === 'online') {
      toast({ title: t('workers.onlineDeleteBlocked'), variant: 'destructive' })
      await loadWorkers()
      return
    }

    toast({ title: t('workers.notFound'), variant: 'destructive' })
    await loadWorkers()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    deletingWorkerId.value = null
  }
}

function workflowHref(usage: WorkerWorkflowUsage): string {
  return workflowVersionEditorPath(orgId.value, usage.workflow_id, usage.workflow_version_id)
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = Math.floor((now - then) / 1000)
  const unit = diff < 60
    ? 'seconds'
    : diff < 3600
      ? 'minutes'
      : diff < 86400
        ? 'hours'
        : 'days'

  const value = unit === 'seconds'
    ? diff
    : unit === 'minutes'
      ? Math.floor(diff / 60)
      : unit === 'hours'
        ? Math.floor(diff / 3600)
        : Math.floor(diff / 86400)

  if (typeof locale.value === 'string' && locale.value.startsWith('ru')) {
    return `${value}${t(`workers.relativeTime.${unit}`)}`
  }

  return t(`workers.relativeTime.${unit}`, { value })
}

const columns: ColumnDef<Worker>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workers.name') }),
    cell: ({ row }) => h('span', { class: 'font-medium' }, row.getValue('name')),
  },
  {
    accessorKey: 'work_type_name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workers.workType') }),
    cell: ({ row }) => h(Badge, { variant: 'outline' }, () => row.getValue('work_type_name') || '-'),
  },
  {
    accessorKey: 'status',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workers.status') }),
    cell: ({ row }) => {
      const status = row.getValue('status') as string
      return h(StatusBadge, { status: status as any })
    },
  },
  {
    accessorKey: 'last_heartbeat_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workers.lastHeartbeat') }),
    cell: ({ row }) => {
      const val = row.getValue('last_heartbeat_at') as string
      return h('span', { class: 'text-muted-foreground text-sm' }, val ? formatRelativeTime(val) : '-')
    },
  },
  {
    id: 'actions',
    header: () => h('span', { class: 'sr-only' }, t('common.actions')),
    cell: ({ row }) => {
      const worker = row.original
      const canDelete = worker.status === 'offline'
      return h(DropdownMenu, {}, {
        default: () => [
          h(DropdownMenuTrigger, { asChild: true }, () =>
            h(Button, { variant: 'ghost', size: 'icon', title: t('common.actions') }, () =>
              h(MoreHorizontal, { class: 'h-4 w-4' }),
            ),
          ),
          h(DropdownMenuContent, { align: 'end' }, () => [
            h(DropdownMenuItem, {
              onClick: () => openWorkflowUsages(worker),
            }, () => [h(Eye, { class: 'mr-2 h-4 w-4' }), t('workers.viewWorkflowUsages')]),
            h(DropdownMenuItem, {
              disabled: !canDelete || deletingWorkerId.value === worker.id,
              title: canDelete ? t('workers.delete') : t('workers.deleteOfflineOnly'),
              onClick: () => onDeleteWorker(worker),
            }, () => [h(Trash2, { class: 'mr-2 h-4 w-4 text-destructive' }), t('workers.delete')]),
          ]),
        ],
      })
    },
    size: 64,
  },
]

onMounted(loadWorkers)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">
        {{ t('workers.title') }}
      </h1>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-3">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>

    <!-- Empty state -->
    <EmptyState
      v-else-if="workers.length === 0"
      :icon="Server"
      :heading="t('empty.workers.heading')"
      :body="t('empty.workers.body')"
    />

    <!-- Data table -->
    <DataTable
      v-else
      :columns="columns"
      :data="workers"
    />

    <div
      v-for="worker in workers"
      :key="`usage-${worker.id}`"
      class="space-y-2"
    >
      <div
        v-if="blockedUsages[worker.id]?.length"
        class="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm"
      >
        <div class="font-medium">
          {{ t('workers.usedByWorkflows', { name: worker.name }) }}
        </div>
        <div class="mt-2 flex flex-wrap gap-2">
          <NuxtLink
            v-for="usage in blockedUsages[worker.id]"
            :key="`${usage.workflow_id}-${usage.workflow_version_id}`"
            :to="workflowHref(usage)"
            class="text-primary underline-offset-4 hover:underline"
          >
            {{ usage.workflow_name }} v{{ usage.workflow_version_number }}
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- Workflow Versions Sheet -->
    <Sheet v-model:open="sheetOpen">
      <SheetContent class="sm:max-w-lg overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{{ selectedWorker?.name }} - {{ t('workers.workflowVersionsTitle') }}</SheetTitle>
          <SheetDescription>
            {{ t('workers.workflowVersionsDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 space-y-4">
          <div v-if="usageLoading" class="space-y-2">
            <Skeleton class="h-12 w-full" />
            <Skeleton class="h-12 w-full" />
          </div>

          <template v-else>
            <div
              v-if="selectedWorkerUsages.length === 0"
              class="text-sm text-muted-foreground text-center py-4"
            >
              {{ t('workers.noWorkflowUsages') }}
            </div>
            <div v-else class="space-y-2">
              <NuxtLink
                v-for="usage in selectedWorkerUsages"
                :key="`${usage.workflow_id}-${usage.workflow_version_id}`"
                :to="workflowHref(usage)"
                class="block rounded-md border p-3 text-sm text-foreground hover:bg-muted"
              >
                <span class="font-medium">{{ usage.workflow_name }}</span>
                <span class="ml-2 text-muted-foreground text-xs">
                  v{{ usage.workflow_version_number }}
                </span>
              </NuxtLink>
            </div>
          </template>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>

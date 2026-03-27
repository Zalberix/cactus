<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { MessageSquare, Pause, Play } from 'lucide-vue-next'
import type { MessageListItem } from '~/composables/useMessages'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import { Button } from '~/components/ui/button'
import { Skeleton } from '~/components/ui/skeleton'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))

const { fetchMessages } = useMessages()

const messages = ref<MessageListItem[]>([])
const loading = ref(true)
const page = ref(1)
const pollingInterval = ref(15000)

async function loadMessages() {
  try {
    const result = await fetchMessages(orgId.value, page.value)
    messages.value = result.data
  }
  catch {
    // Silent fail for polling -- initial load handles error display
  }
  finally {
    loading.value = false
  }
}

const { isPolling, pause, resume } = usePolling(
  loadMessages,
  pollingInterval,
)

function togglePolling() {
  if (isPolling.value) {
    pause()
  }
  else {
    resume()
  }
}

function viewMessage(messageId: number) {
  router.push(`/org/${orgId.value}/messages/${messageId}`)
}

const columns: ColumnDef<MessageListItem>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.id') }),
    cell: ({ row }) => h('span', { class: 'font-mono text-sm' }, `#${row.getValue('id')}`),
    size: 80,
  },
  {
    accessorKey: 'workflow_name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.workflow') }),
    cell: ({ row }) => h('span', { class: 'font-medium' }, row.getValue('workflow_name') || '-'),
  },
  {
    accessorKey: 'status',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.status') }),
    cell: ({ row }) => {
      const status = row.getValue('status') as string
      const mapped = status === 'completed' ? 'done' : status === 'failed' ? 'error' : status
      return h(StatusBadge, { status: mapped as any })
    },
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('messages.createdAt') }),
    cell: ({ row }) => {
      const val = row.getValue('created_at') as string
      return h('span', { class: 'text-sm text-muted-foreground' }, val ? new Date(val).toLocaleString() : '-')
    },
  },
  {
    id: 'actions',
    header: () => h('span', { class: 'sr-only' }, t('common.actions')),
    cell: ({ row }) => {
      return h(Button, {
        variant: 'ghost',
        size: 'sm',
        onClick: () => viewMessage(row.original.id),
      }, () => t('messages.viewDetails'))
    },
    size: 120,
  },
]

onMounted(loadMessages)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">
        {{ t('messages.title') }}
      </h1>

      <!-- Polling indicator -->
      <div class="flex items-center gap-2 text-sm text-muted-foreground">
        <span>{{ t('messages.autoRefresh', { seconds: pollingInterval / 1000 }) }}</span>
        <Button
          variant="ghost"
          size="icon"
          class="size-8"
          @click="togglePolling"
        >
          <Pause v-if="isPolling" class="size-4" />
          <Play v-else class="size-4" />
        </Button>
      </div>
    </div>

    <!-- Stub banner -->
    <div class="rounded-md border border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950 px-4 py-3 text-sm text-amber-800 dark:text-amber-200">
      {{ t('messages.stubBanner') }}
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-3">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>

    <!-- Empty state -->
    <EmptyState
      v-else-if="messages.length === 0"
      :icon="MessageSquare"
      :heading="t('empty.messages.heading')"
      :body="t('empty.messages.body')"
    />

    <!-- Data table -->
    <DataTable
      v-else
      :columns="columns"
      :data="messages"
    />
  </div>
</template>

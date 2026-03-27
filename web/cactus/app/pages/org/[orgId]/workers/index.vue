<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Server, Eye } from 'lucide-vue-next'
import type { Worker, SettingsRevision } from '~/composables/useWorkers'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import DynamicSettingsForm from '~/components/forms/DynamicSettingsForm.vue'
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
import { Separator } from '~/components/ui/separator'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const orgId = computed(() => Number(route.params.orgId))

const { fetchWorkersForOrg, fetchRevisions, createRevision, fetchSchema } = useWorkers()

const workers = ref<Worker[]>([])
const loading = ref(true)

const sheetOpen = ref(false)
const selectedWorker = ref<Worker | null>(null)
const revisions = ref<SettingsRevision[]>([])
const revisionsLoading = ref(false)
const showCreateForm = ref(false)
const formData = ref<Record<string, unknown>>({})
const submitting = ref(false)

const settingsSchema = ref<Record<string, unknown>>({})
const schemaLoading = ref(false)

async function loadWorkers() {
  loading.value = true
  try {
    workers.value = await fetchWorkersForOrg(orgId.value)
  }
  catch (err) {
    toast({
      title: t('error.server'),
      variant: 'destructive',
    })
  }
  finally {
    loading.value = false
  }
}

async function openRevisions(worker: Worker) {
  selectedWorker.value = worker
  sheetOpen.value = false
  showCreateForm.value = false
  formData.value = {}

  await nextTick()
  sheetOpen.value = true

  if (!worker.schema_id) return

  // Fetch schema definition for DynamicSettingsForm
  schemaLoading.value = true
  try {
    const schema = await fetchSchema(worker.schema_id)
    settingsSchema.value = schema.settings_schema ?? {}
  }
  catch {
    settingsSchema.value = {}
  }
  finally {
    schemaLoading.value = false
  }

  revisionsLoading.value = true
  try {
    revisions.value = await fetchRevisions(worker.schema_id)
  }
  catch {
    revisions.value = []
  }
  finally {
    revisionsLoading.value = false
  }
}

async function onSubmitRevision() {
  if (!selectedWorker.value?.schema_id) return

  submitting.value = true
  try {
    await createRevision(selectedWorker.value.schema_id, formData.value)
    toast({ title: t('workers.revisionCreated') })
    showCreateForm.value = false
    formData.value = {}
    revisions.value = await fetchRevisions(selectedWorker.value.schema_id)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = Math.floor((now - then) / 1000)

  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
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
      return h(Button, {
        variant: 'ghost',
        size: 'sm',
        onClick: () => openRevisions(row.original),
      }, () => [
        h(Eye, { class: 'size-4 mr-1' }),
        t('workers.viewRevisions'),
      ])
    },
    size: 160,
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

    <!-- Settings Revisions Sheet -->
    <Sheet v-model:open="sheetOpen">
      <SheetContent class="sm:max-w-lg overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{{ selectedWorker?.name }} - {{ t('workers.settingsRevisions') }}</SheetTitle>
          <SheetDescription>
            {{ t('workers.revisionsDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 space-y-4">
          <!-- Schema ID info -->
          <div v-if="selectedWorker?.schema_id" class="text-sm text-muted-foreground">
            Schema ID: {{ selectedWorker.schema_id }}
          </div>
          <div v-else class="text-sm text-muted-foreground">
            {{ t('workers.noSchema') }}
          </div>

          <!-- Revisions list -->
          <template v-if="selectedWorker?.schema_id">
            <div v-if="revisionsLoading" class="space-y-2">
              <Skeleton class="h-12 w-full" />
              <Skeleton class="h-12 w-full" />
            </div>

            <template v-else>
              <div
                v-for="rev in revisions"
                :key="rev.id"
                class="rounded-md border p-3 text-sm space-y-1"
              >
                <div class="flex items-center justify-between">
                  <span class="font-medium">Revision #{{ rev.id }}</span>
                  <span class="text-xs text-muted-foreground">
                    {{ new Date(rev.created_at).toLocaleString() }}
                  </span>
                </div>
                <pre class="text-xs bg-muted p-2 rounded overflow-x-auto">{{ JSON.stringify(rev.settings_data, null, 2) }}</pre>
              </div>

              <div v-if="revisions.length === 0" class="text-sm text-muted-foreground text-center py-4">
                {{ t('workers.noRevisions') }}
              </div>
            </template>

            <Separator />

            <!-- Create revision form -->
            <div v-if="!showCreateForm">
              <Button size="sm" @click="showCreateForm = true">
                {{ t('workers.createRevision') }}
              </Button>
            </div>

            <div v-else class="space-y-4">
              <h4 class="font-medium text-sm">{{ t('workers.createRevision') }}</h4>

              <DynamicSettingsForm
                v-if="Object.keys(settingsSchema).length > 0"
                :schema="settingsSchema"
                v-model="formData"
              />

              <!-- Fallback: raw JSON input when no schema available -->
              <div v-else class="space-y-2">
                <label class="text-sm font-medium">{{ t('workers.settingsJson') }}</label>
                <textarea
                  :value="JSON.stringify(formData, null, 2)"
                  class="w-full h-32 rounded-md border bg-background px-3 py-2 text-sm font-mono"
                  @input="(e: Event) => {
                    try {
                      formData = JSON.parse((e.target as HTMLTextAreaElement).value)
                    } catch {}
                  }"
                />
              </div>

              <div class="flex gap-2">
                <Button
                  size="sm"
                  :disabled="submitting"
                  @click="onSubmitRevision"
                >
                  {{ submitting ? t('common.loading') : t('common.save') }}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  @click="showCreateForm = false"
                >
                  {{ t('common.cancel') }}
                </Button>
              </div>
            </div>
          </template>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>

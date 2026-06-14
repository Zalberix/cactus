<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Ban, KeyRound, Plus, RefreshCw } from 'lucide-vue-next'
import type {
  CreateWorkerBootstrapTokenInput,
  WorkerBootstrapToken,
  WorkType,
} from '~/composables/useWorkers'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Checkbox } from '~/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'
import { toast } from '~/components/ui/toast/use-toast'

interface TokenForm {
  name: string
  work_type_id: number
  description: string
  expires_at: string
  max_active_workers: number
}

const { t } = useI18n()
const route = useRoute()
const orgId = computed(() => Number(route.params.orgId))

const {
  fetchWorkTypes,
  fetchWorkerBootstrapTokens,
  createWorkerBootstrapToken,
  revokeWorkerBootstrapToken,
} = useWorkers()

const tokens = ref<WorkerBootstrapToken[]>([])
const workTypes = ref<WorkType[]>([])
const loading = ref(true)
const submitting = ref(false)
const createOpen = ref(false)
const showOnce = ref(false)
const plaintext = ref('')
const revokeOpen = ref(false)
const revokeActiveSessions = ref(true)
const tokenToRevoke = ref<WorkerBootstrapToken | null>(null)

const form = ref<TokenForm>({
  name: '',
  work_type_id: 0,
  description: '',
  expires_at: '',
  max_active_workers: 1,
})

const workTypeSelectValue = computed({
  get: () => String(form.value.work_type_id || 0),
  set: (value: string) => {
    form.value.work_type_id = Number(value)
  },
})

const columns: ColumnDef<WorkerBootstrapToken>[] = [
  {
    id: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workerTokens.name') }),
    cell: ({ row }) => h('div', { class: 'space-y-1' }, [
      h('span', { class: 'font-medium' }, row.original.name),
      row.original.description
        ? h('p', { class: 'max-w-[360px] truncate text-xs text-muted-foreground' }, row.original.description)
        : null,
    ]),
  },
  {
    id: 'work_type_id',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workerTokens.workType') }),
    cell: ({ row }) => workTypeName(row.original.work_type_id),
  },
  {
    accessorKey: 'status',
    header: t('workerTokens.status'),
    cell: ({ row }) => {
      const active = row.original.status === 'active'
      return h(Badge, {
        variant: active ? 'default' : 'secondary',
        class: active ? 'bg-green-600 hover:bg-green-700' : '',
      }, () => active ? t('status.active') : t('status.inactive'))
    },
  },
  {
    id: 'active_workers',
    header: t('workerTokens.activeWorkers'),
    cell: ({ row }) => `${row.original.active_worker_count}/${row.original.max_active_workers}`,
  },
  {
    id: 'expires_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workerTokens.expiresAt') }),
    cell: ({ row }) => formatDate(row.original.expires_at),
  },
  {
    id: 'last_used_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workerTokens.lastUsedAt') }),
    cell: ({ row }) => formatDate(row.original.last_used_at),
  },
  {
    id: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workerTokens.createdAt') }),
    cell: ({ row }) => formatDate(row.original.created_at),
  },
  {
    id: 'actions',
    header: '',
    size: 120,
    cell: ({ row }) => h('div', { class: 'flex justify-end' }, [
      h(Button, {
        variant: 'ghost',
        size: 'sm',
        class: 'gap-2 text-destructive hover:text-destructive',
        disabled: row.original.status !== 'active',
        onClick: () => openRevokeDialog(row.original),
      }, () => [h(Ban, { class: 'h-4 w-4' }), t('workerTokens.revoke')]),
    ]),
  },
]

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString()
}

function workTypeName(workTypeId: number) {
  return workTypes.value.find(workType => workType.id === workTypeId)?.name ?? `${workTypeId}`
}

function resetForm() {
  form.value = {
    name: '',
    work_type_id: workTypes.value[0]?.id ?? 0,
    description: '',
    expires_at: '',
    max_active_workers: 1,
  }
}

async function loadTokens() {
  tokens.value = await fetchWorkerBootstrapTokens(orgId.value)
}

async function loadData() {
  loading.value = true
  try {
    const [loadedWorkTypes] = await Promise.all([
      fetchWorkTypes(),
      loadTokens(),
    ])
    workTypes.value = loadedWorkTypes
    if (!form.value.work_type_id) {
      form.value.work_type_id = loadedWorkTypes[0]?.id ?? 0
    }
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function openCreateDialog() {
  resetForm()
  plaintext.value = ''
  showOnce.value = false
  createOpen.value = true
}

async function onCreateToken() {
  submitting.value = true
  try {
    const input: CreateWorkerBootstrapTokenInput = {
      name: form.value.name.trim(),
      work_type_id: Number(form.value.work_type_id),
      max_active_workers: Number(form.value.max_active_workers),
    }
    const description = form.value.description.trim()
    if (description) input.description = description
    if (form.value.expires_at) {
      input.expires_at = new Date(form.value.expires_at).toISOString()
    }

    const result = await createWorkerBootstrapToken(orgId.value, input)
    plaintext.value = result.plaintext
    showOnce.value = true
    toast({ title: t('workerTokens.created') })
    await loadTokens()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function closeShowOnce() {
  createOpen.value = false
  showOnce.value = false
  plaintext.value = ''
}

function openRevokeDialog(token: WorkerBootstrapToken) {
  tokenToRevoke.value = token
  revokeActiveSessions.value = true
  revokeOpen.value = true
}

async function onRevokeToken() {
  if (!tokenToRevoke.value) return
  submitting.value = true
  try {
    await revokeWorkerBootstrapToken(tokenToRevoke.value.id, revokeActiveSessions.value)
    toast({ title: t('workerTokens.revoked') })
    revokeOpen.value = false
    tokenToRevoke.value = null
    await loadTokens()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold">
          {{ t('workerTokens.title') }}
        </h2>
        <p class="text-sm text-muted-foreground">
          {{ t('workerTokens.description') }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Button variant="outline" size="icon" :disabled="loading" :title="t('error.retry')" @click="loadData">
          <RefreshCw class="h-4 w-4" />
        </Button>
        <Button class="gap-2" @click="openCreateDialog">
          <Plus class="h-4 w-4" />
          {{ t('workerTokens.create') }}
        </Button>
      </div>
    </div>

    <DataTable
      v-if="!loading || tokens.length > 0"
      :columns="columns"
      :data="tokens"
      :loading="loading"
    >
      <template #empty>
        <EmptyState
          :icon="KeyRound"
          :heading="t('workerTokens.emptyTitle')"
          :body="t('workerTokens.emptyBody')"
          :cta-label="t('workerTokens.create')"
          @cta="openCreateDialog"
        />
      </template>
    </DataTable>

    <DataTable
      v-if="loading && tokens.length === 0"
      :columns="columns"
      :data="[]"
      :loading="true"
    />

    <Dialog v-model:open="createOpen">
      <DialogContent @interact-outside.prevent>
        <template v-if="!showOnce">
          <DialogHeader>
            <DialogTitle>{{ t('workerTokens.create') }}</DialogTitle>
            <DialogDescription>{{ t('workerTokens.createDescription') }}</DialogDescription>
          </DialogHeader>

          <div class="space-y-4">
            <div class="space-y-2">
              <Label for="worker-token-name">{{ t('workerTokens.name') }}</Label>
              <Input
                id="worker-token-name"
                v-model="form.name"
                :placeholder="t('workerTokens.namePlaceholder')"
              />
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <div class="space-y-2">
                <Label for="worker-token-work-type">{{ t('workerTokens.workType') }}</Label>
                <Select v-model="workTypeSelectValue">
                  <SelectTrigger id="worker-token-work-type" class="h-10 w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0" disabled>
                      {{ t('workerTokens.selectWorkType') }}
                    </SelectItem>
                    <SelectItem v-for="workType in workTypes" :key="workType.id" :value="String(workType.id)">
                      {{ workType.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div class="space-y-2">
                <Label for="worker-token-max">{{ t('workerTokens.maxActiveWorkers') }}</Label>
                <Input
                  id="worker-token-max"
                  v-model.number="form.max_active_workers"
                  min="1"
                  type="number"
                />
              </div>
            </div>

            <div class="space-y-2">
              <Label for="worker-token-expires">{{ t('workerTokens.expiresAt') }}</Label>
              <Input
                id="worker-token-expires"
                v-model="form.expires_at"
                type="datetime-local"
              />
            </div>

            <div class="space-y-2">
              <Label for="worker-token-description">{{ t('workerTokens.descriptionField') }}</Label>
              <textarea
                id="worker-token-description"
                v-model="form.description"
                class="min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                :placeholder="t('workerTokens.descriptionPlaceholder')"
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="createOpen = false">
              {{ t('common.cancel') }}
            </Button>
            <Button
              :disabled="submitting || !form.name.trim() || !form.work_type_id || form.max_active_workers < 1"
              @click="onCreateToken"
            >
              {{ t('workerTokens.create') }}
            </Button>
          </DialogFooter>
        </template>

        <template v-else>
          <DialogHeader>
            <DialogTitle>{{ t('workerTokens.created') }}</DialogTitle>
            <DialogDescription>{{ t('workerTokens.plaintextWarning') }}</DialogDescription>
          </DialogHeader>

          <div class="space-y-3">
            <Label>{{ t('workerTokens.token') }}</Label>
            <code class="block w-full rounded-md border bg-muted px-3 py-2 text-sm font-mono break-all">
              {{ plaintext }}
            </code>
          </div>

          <DialogFooter>
            <Button @click="closeShowOnce">
              {{ t('common.cancel') }}
            </Button>
          </DialogFooter>
        </template>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="revokeOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('workerTokens.revoke') }}</DialogTitle>
          <DialogDescription>{{ t('workerTokens.revokeDescription') }}</DialogDescription>
        </DialogHeader>

        <label class="flex cursor-pointer items-start gap-3 rounded-md border px-3 py-2 text-sm">
          <Checkbox
            v-model="revokeActiveSessions"
          />
          <span>{{ t('workerTokens.revokeActiveSessions') }}</span>
        </label>

        <DialogFooter>
          <Button type="button" variant="outline" @click="revokeOpen = false">
            {{ t('common.cancel') }}
          </Button>
          <Button variant="destructive" :disabled="submitting" @click="onRevokeToken">
            {{ t('workerTokens.revoke') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

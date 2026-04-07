<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Key, ArrowLeft, Download, ShieldCheck, ShieldOff } from 'lucide-vue-next'
import type { Token, CreateTokenResponse } from '~/composables/useSystems'
import DataTable from '~/components/tables/DataTable.vue'
import DataTableColumnHeader from '~/components/tables/DataTableColumnHeader.vue'
import EmptyState from '~/components/feedback/EmptyState.vue'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
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
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))
const systemId = computed(() => Number(route.params.systemId))

const { fetchSystems, fetchTokens, createToken, revokeToken, activateToken } = useSystems()

const tokens = ref<Token[]>([])
const systemName = ref('')
const loading = ref(true)

// Create token dialog state
const createOpen = ref(false)
const tokenName = ref('')
const submitting = ref(false)

// Show-once state
const showOnce = ref(false)
const createdToken = ref<CreateTokenResponse | null>(null)

// Revoke/Activate dialog state
const confirmOpen = ref(false)
const confirmAction = ref<'revoke' | 'activate'>('revoke')
const tokenToAction = ref<Token | null>(null)

const columns: ColumnDef<Token>[] = [
  {
    id: 'name',
    header: t('tokens.name'),
    cell: ({ row }) => {
      const name = row.original.name
      return h('span', { class: 'font-medium' }, name || '-')
    },
  },
  {
    accessorKey: 'public_token',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('tokens.publicToken') }),
    cell: ({ row }) => h('code', { class: 'text-sm font-mono bg-muted px-2 py-1 rounded' }, row.getValue('public_token')),
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('tokens.createdAt') }),
    cell: ({ row }) => {
      const date = new Date(row.getValue('created_at') as string)
      return h('span', { class: 'text-muted-foreground' }, date.toLocaleDateString())
    },
  },
  {
    id: 'status',
    header: t('tokens.status'),
    cell: ({ row }) => {
      const active = row.original.is_active
      return h(Badge, {
        variant: active ? 'default' : 'secondary',
        class: active ? 'bg-green-600 hover:bg-green-700' : '',
      }, () => active ? t('status.active') : t('status.inactive'))
    },
  },
  {
    id: 'actions',
    header: '',
    size: 120,
    cell: ({ row }) => {
      const token = row.original
      if (token.is_active) {
        return h(Button, {
          variant: 'ghost',
          size: 'sm',
          class: 'text-destructive hover:text-destructive',
          onClick: () => onConfirmAction(token, 'revoke'),
        }, () => t('tokens.revoke'))
      }
      return h(Button, {
        variant: 'ghost',
        size: 'sm',
        onClick: () => onConfirmAction(token, 'activate'),
      }, () => t('tokens.activate'))
    },
  },
]

async function loadData() {
  loading.value = true
  try {
    const systems = await fetchSystems(orgId.value)
    const system = systems.find(s => s.id === systemId.value)
    systemName.value = system?.name ?? ''
    tokens.value = await fetchTokens(systemId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function openCreateDialog() {
  tokenName.value = ''
  showOnce.value = false
  createdToken.value = null
  createOpen.value = true
}

async function onCreateToken() {
  submitting.value = true
  try {
    const result = await createToken(systemId.value, tokenName.value)
    createdToken.value = result
    showOnce.value = true
    fetchTokens(systemId.value).then(data => tokens.value = data).catch(() => {})
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function downloadTokenJson() {
  if (!createdToken.value) return
  const data = {
    public_token: createdToken.value.public_token,
    private_token: createdToken.value.private_token,
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `token-${createdToken.value.public_token.slice(0, 8)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function onCloseShowOnce() {
  createOpen.value = false
  showOnce.value = false
  createdToken.value = null
}

function onConfirmAction(token: Token, action: 'revoke' | 'activate') {
  tokenToAction.value = token
  confirmAction.value = action
  confirmOpen.value = true
}

async function onExecuteAction() {
  if (!tokenToAction.value) return
  submitting.value = true
  try {
    if (confirmAction.value === 'revoke') {
      await revokeToken(tokenToAction.value.id)
      toast({ title: t('tokens.revoked') })
    }
    else {
      await activateToken(tokenToAction.value.id)
      toast({ title: t('tokens.activated') })
    }
    confirmOpen.value = false
    tokenToAction.value = null
    tokens.value = await fetchTokens(systemId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function goBack() {
  router.push(`/org/${orgId.value}/settings/systems`)
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Back + Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <Button variant="ghost" size="sm" class="gap-2" @click="goBack">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <h2 v-if="systemName" class="text-lg font-semibold">
          {{ systemName }} - {{ t('tokens.title') }}
        </h2>
      </div>
      <Button @click="openCreateDialog">
        {{ t('tokens.create') }}
      </Button>
    </div>

    <!-- Data Table -->
    <DataTable
      v-if="!loading || tokens.length > 0"
      :columns="columns"
      :data="tokens"
      :loading="loading"
    >
      <template #empty>
        <EmptyState
          :icon="Key"
          :heading="t('empty.tokens.heading')"
          :body="t('empty.tokens.body')"
          :cta-label="t('empty.tokens.cta')"
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

    <!-- Create Token / Show-Once Dialog -->
    <Dialog v-model:open="createOpen">
      <DialogContent @interact-outside.prevent>
        <!-- Step 1: Create form -->
        <template v-if="!showOnce">
          <DialogHeader>
            <DialogTitle>{{ t('tokens.create') }}</DialogTitle>
            <DialogDescription>{{ t('tokens.createDescription') }}</DialogDescription>
          </DialogHeader>
          <div class="space-y-4">
            <div class="space-y-2">
              <Label>{{ t('tokens.name') }}</Label>
              <Input
                v-model="tokenName"
                :placeholder="t('tokens.namePlaceholder')"
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" @click="createOpen = false">
              {{ t('common.cancel') }}
            </Button>
            <Button :disabled="submitting || !tokenName.trim()" @click="onCreateToken">
              {{ t('tokens.create') }}
            </Button>
          </DialogFooter>
        </template>

        <!-- Step 2: Show-once screen -->
        <template v-else>
          <DialogHeader>
            <DialogTitle>{{ t('tokenShowOnce.heading') }}</DialogTitle>
            <DialogDescription>{{ t('tokenShowOnce.body') }}</DialogDescription>
          </DialogHeader>

          <div class="space-y-4">
            <!-- Public Token -->
            <div class="space-y-2">
              <Label>{{ t('tokens.publicToken') }}</Label>
              <code class="block w-full rounded-md border bg-muted px-3 py-2 text-sm font-mono break-all">
                {{ createdToken?.public_token }}
              </code>
            </div>

            <!-- Private Token -->
            <div class="space-y-2">
              <Label>{{ t('tokens.privateToken') }}</Label>
              <code class="block w-full rounded-md border bg-muted px-3 py-2 text-sm font-mono break-all">
                {{ createdToken?.private_token }}
              </code>
            </div>
          </div>

          <div class="rounded-md border border-yellow-500/50 bg-yellow-500/10 p-3 text-sm text-yellow-700 dark:text-yellow-400">
            {{ t('tokens.closeWarning') }}
          </div>

          <DialogFooter class="gap-2 sm:gap-0">
            <Button variant="outline" class="gap-2" @click="downloadTokenJson">
              <Download class="h-4 w-4" />
              {{ t('tokens.downloadJson') }}
            </Button>
            <Button @click="onCloseShowOnce">
              {{ t('common.cancel') }}
            </Button>
          </DialogFooter>
        </template>
      </DialogContent>
    </Dialog>

    <!-- Revoke / Activate Confirmation Dialog -->
    <Dialog v-model:open="confirmOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {{ confirmAction === 'revoke' ? t('destructive.revokeToken.title') : t('destructive.activateToken.title') }}
          </DialogTitle>
          <DialogDescription>
            {{ confirmAction === 'revoke' ? t('destructive.revokeToken.body') : t('destructive.activateToken.body') }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" @click="confirmOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button
            :variant="confirmAction === 'revoke' ? 'destructive' : 'default'"
            :disabled="submitting"
            @click="onExecuteAction"
          >
            {{ confirmAction === 'revoke' ? t('destructive.revokeToken.confirm') : t('destructive.activateToken.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

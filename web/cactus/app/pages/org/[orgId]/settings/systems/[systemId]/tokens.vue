<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Key, ArrowLeft, Copy, Check } from 'lucide-vue-next'
import { useClipboard } from '@vueuse/core'
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

const { fetchSystems, fetchTokens, createToken, revokeToken } = useSystems()

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
const { copy, copied } = useClipboard({ copiedDuring: 2000 })

// Revoke dialog state
const revokeOpen = ref(false)
const tokenToRevoke = ref<Token | null>(null)

const columns: ColumnDef<Token>[] = [
  {
    accessorKey: 'public_token',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('tokens.publicToken') }),
    cell: ({ row }) => h('code', { class: 'text-sm font-mono bg-muted px-2 py-1 rounded' }, row.getValue('public_token')),
  },
  {
    id: 'name',
    header: t('tokens.name'),
    cell: ({ row }) => {
      const name = row.original.name
      return h('span', { class: 'text-muted-foreground' }, name || '-')
    },
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
    size: 80,
    cell: ({ row }) => {
      if (!row.original.is_active) return null
      return h(Button, {
        variant: 'ghost',
        size: 'sm',
        class: 'text-destructive hover:text-destructive',
        onClick: () => onConfirmRevoke(row.original),
      }, () => t('destructive.revokeToken.confirm'))
    },
  },
]

async function loadData() {
  loading.value = true
  try {
    // Load system name
    const systems = await fetchSystems(orgId.value)
    const system = systems.find(s => s.id === systemId.value)
    systemName.value = system?.name ?? ''

    // Load tokens
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
    const result = await createToken(systemId.value, tokenName.value || undefined)
    createdToken.value = result
    showOnce.value = true
    // Reload tokens list in background
    fetchTokens(systemId.value).then(data => tokens.value = data).catch(() => {})
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function onCopyPrivateToken() {
  if (createdToken.value) {
    copy(createdToken.value.private_token)
  }
}

function onCloseShowOnce() {
  createOpen.value = false
  showOnce.value = false
  createdToken.value = null
}

function onConfirmRevoke(token: Token) {
  tokenToRevoke.value = token
  revokeOpen.value = true
}

async function onRevoke() {
  if (!tokenToRevoke.value) return
  submitting.value = true
  try {
    await revokeToken(tokenToRevoke.value.id)
    revokeOpen.value = false
    tokenToRevoke.value = null
    toast({ title: t('tokens.revoked') })
    await loadData()
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
            <Button :disabled="submitting" @click="onCreateToken">
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
              <div class="flex gap-2">
                <code class="flex-1 rounded-md border bg-muted px-3 py-2 text-sm font-mono break-all">
                  {{ createdToken?.private_token }}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  class="shrink-0 gap-2"
                  @click="onCopyPrivateToken"
                >
                  <template v-if="copied">
                    <Check class="h-4 w-4" />
                    {{ t('tokenShowOnce.copied') }}
                  </template>
                  <template v-else>
                    <Copy class="h-4 w-4" />
                    {{ t('tokenShowOnce.copyButton') }}
                  </template>
                </Button>
              </div>
            </div>
          </div>

          <div class="rounded-md border border-yellow-500/50 bg-yellow-500/10 p-3 text-sm text-yellow-700 dark:text-yellow-400">
            {{ t('tokens.closeWarning') }}
          </div>

          <DialogFooter>
            <Button @click="onCloseShowOnce">
              {{ t('tokens.confirmClose') }}
            </Button>
          </DialogFooter>
        </template>
      </DialogContent>
    </Dialog>

    <!-- Revoke Confirmation Dialog -->
    <Dialog v-model:open="revokeOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('destructive.revokeToken.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.revokeToken.body') }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" @click="revokeOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button
            variant="destructive"
            :disabled="submitting"
            @click="onRevoke"
          >
            {{ t('destructive.revokeToken.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

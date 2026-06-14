<script setup lang="ts">
import { Download } from 'lucide-vue-next'
import type { CreateTokenResponse, Token } from '~/composables/useSystems'
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
import { toast } from '~/components/ui/toast/use-toast'

const props = defineProps<{
  workflowId: number
  systemId: number | null
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const { t } = useI18n()
const {
  fetchTokens,
  createToken,
  fetchWorkflowTokens,
  bindTokenWorkflow,
  unbindTokenWorkflow,
} = useSystems()

const tokens = ref<Token[]>([])
const originalTokenIds = ref<Set<number>>(new Set())
const selectedTokenIds = ref<Set<number>>(new Set())
const loading = ref(false)
const saving = ref(false)
const tokenCreateName = ref('')
const tokenCreating = ref(false)
const createdToken = ref<CreateTokenResponse | null>(null)

watch(
  () => props.open,
  async (open) => {
    if (open) await loadTokens()
  },
)

async function loadTokens() {
  if (!props.systemId) return
  loading.value = true
  createdToken.value = null
  tokenCreateName.value = ''
  try {
    const [allTokens, links] = await Promise.all([
      fetchTokens(props.systemId),
      fetchWorkflowTokens(props.workflowId),
    ])
    tokens.value = allTokens
    const linkedIds = new Set(links.map(link => link.system_token_id))
    originalTokenIds.value = new Set(linkedIds)
    selectedTokenIds.value = new Set(linkedIds)
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function isTokenSelected(tokenId: number) {
  return selectedTokenIds.value.has(tokenId)
}

function toggleToken(tokenId: number) {
  const next = new Set(selectedTokenIds.value)
  if (next.has(tokenId)) next.delete(tokenId)
  else next.add(tokenId)
  selectedTokenIds.value = next
}

async function saveAccess() {
  saving.value = true
  try {
    const selected = selectedTokenIds.value
    const original = originalTokenIds.value
    await Promise.all([
      ...[...selected].filter(id => !original.has(id)).map(id => bindTokenWorkflow(id, props.workflowId)),
      ...[...original].filter(id => !selected.has(id)).map(id => unbindTokenWorkflow(id, props.workflowId)),
    ])
    isOpen.value = false
    toast({ title: t('tokens.workflowAccessSaved') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

async function createTokenForWorkflow() {
  if (!props.systemId || !tokenCreateName.value.trim()) return
  tokenCreating.value = true
  try {
    const token = await createToken(props.systemId, tokenCreateName.value.trim())
    await bindTokenWorkflow(token.id, props.workflowId)
    createdToken.value = token
    await loadTokens()
    selectedTokenIds.value = new Set([...selectedTokenIds.value, token.id])
    tokenCreateName.value = ''
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    tokenCreating.value = false
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
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ t('tokens.workflowDialogTitle') }}</DialogTitle>
        <DialogDescription>{{ t('tokens.workflowDialogDescription') }}</DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div class="space-y-2">
          <Label>{{ t('tokens.createForWorkflow') }}</Label>
          <div class="flex gap-2">
            <Input
              v-model="tokenCreateName"
              :placeholder="t('tokens.namePlaceholder')"
              :disabled="tokenCreating"
            />
            <Button
              :disabled="tokenCreating || !tokenCreateName.trim()"
              @click="createTokenForWorkflow"
            >
              {{ t('common.create') }}
            </Button>
          </div>
        </div>

        <div v-if="createdToken" class="space-y-3 rounded-md border border-yellow-500/50 bg-yellow-500/10 p-3">
          <p class="text-sm font-medium">{{ t('tokenShowOnce.heading') }}</p>
          <code class="block rounded-md border bg-background px-3 py-2 text-xs font-mono break-all">
            {{ createdToken.public_token }}
          </code>
          <code class="block rounded-md border bg-background px-3 py-2 text-xs font-mono break-all">
            {{ createdToken.private_token }}
          </code>
          <Button variant="outline" size="sm" class="gap-2" @click="downloadTokenJson">
            <Download class="h-4 w-4" />
            {{ t('tokens.downloadJson') }}
          </Button>
        </div>

        <div class="max-h-[320px] space-y-2 overflow-y-auto pr-1">
          <p v-if="loading" class="text-sm text-muted-foreground">{{ t('common.loading') }}</p>
          <p v-else-if="tokens.length === 0" class="text-sm text-muted-foreground">{{ t('tokens.noTokens') }}</p>
          <template v-else>
            <label
              v-for="token in tokens"
              :key="token.id"
              class="flex cursor-pointer items-center gap-3 rounded-md border px-3 py-2 text-sm hover:bg-muted/50"
            >
              <Checkbox
                :model-value="isTokenSelected(token.id)"
                @update:model-value="toggleToken(token.id)"
              />
              <span class="min-w-0 flex-1">
                <span class="block font-medium">{{ token.name || token.public_token }}</span>
                <span class="block truncate text-xs text-muted-foreground">{{ token.public_token }}</span>
              </span>
              <span class="text-xs" :class="token.is_active ? 'text-green-600' : 'text-muted-foreground'">
                {{ token.is_active ? t('status.active') : t('status.inactive') }}
              </span>
            </label>
          </template>
        </div>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" @click="isOpen = false">{{ t('common.cancel') }}</Button>
        <Button :disabled="loading || saving" @click="saveAccess">{{ t('common.save') }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

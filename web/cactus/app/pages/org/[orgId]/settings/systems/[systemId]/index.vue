<script setup lang="ts">
import { ArrowLeft, Copy, Plus, Trash2, ShieldCheck, ShieldOff } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { System, Token, CreateTokenResponse } from '~/composables/useSystems'
import { Button } from '~/components/ui/button'
import { Badge } from '~/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '~/components/ui/form'
import { Input } from '~/components/ui/input'
import { Separator } from '~/components/ui/separator'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))
const systemId = computed(() => Number(route.params.systemId))

const { fetchSystems, updateSystem, fetchTokens, createToken, revokeToken } = useSystems()

const system = ref<System | null>(null)
const tokens = ref<Token[]>([])
const loading = ref(true)
const saving = ref(false)
const createTokenOpen = ref(false)
const revokeOpen = ref(false)
const tokenToRevoke = ref<Token | null>(null)
const newToken = ref<CreateTokenResponse | null>(null)
const creatingToken = ref(false)
const tokenName = ref('')

const systemSchema = toTypedSchema(
  z.object({
    name: z.string().min(2, t('systems.validation.nameMin')),
    description: z.string().optional(),
  }),
)

const initialValues = computed(() => ({
  name: system.value?.name ?? '',
  description: system.value?.description ?? '',
}))

async function loadData() {
  loading.value = true
  try {
    const [systems, toks] = await Promise.all([
      fetchSystems(orgId.value),
      fetchTokens(systemId.value),
    ])
    system.value = systems.find(s => s.id === systemId.value) ?? null
    tokens.value = toks
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function onSave(values: Record<string, unknown>) {
  saving.value = true
  try {
    await updateSystem(systemId.value, values.name as string, values.description as string)
    toast({ title: t('systems.updated') })
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

async function onCreateToken() {
  creatingToken.value = true
  try {
    newToken.value = await createToken(systemId.value, tokenName.value || undefined)
    tokenName.value = ''
    toast({ title: t('tokens.created') })
    tokens.value = await fetchTokens(systemId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    creatingToken.value = false
  }
}

function onConfirmRevoke(token: Token) {
  tokenToRevoke.value = token
  revokeOpen.value = true
}

async function onRevoke() {
  if (!tokenToRevoke.value) return
  try {
    await revokeToken(tokenToRevoke.value.id)
    revokeOpen.value = false
    tokenToRevoke.value = null
    toast({ title: t('tokens.revoked') })
    tokens.value = await fetchTokens(systemId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  toast({ title: t('common.copied') })
}

function goBack() {
  router.push(`/org/${orgId.value}/settings/systems`)
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6">
    <Button variant="ghost" size="sm" class="gap-2" @click="goBack">
      <ArrowLeft class="h-4 w-4" />
      {{ t('systems.backToSystems') }}
    </Button>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <span class="text-muted-foreground">{{ t('common.loading') }}</span>
    </div>

    <template v-else-if="system">
      <!-- System Details -->
      <Card>
        <CardHeader>
          <CardTitle>{{ t('systems.editSystem') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <Form
            v-slot="{ handleSubmit }"
            :validation-schema="systemSchema"
            :initial-values="initialValues"
            class="space-y-4"
          >
            <form @submit="handleSubmit($event, onSave)">
              <div class="space-y-4">
                <FormField v-slot="{ componentField }" name="name">
                  <FormItem>
                    <FormLabel>{{ t('systems.name') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>

                <FormField v-slot="{ componentField }" name="description">
                  <FormItem>
                    <FormLabel>{{ t('systems.description') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>

              <div class="mt-4 flex justify-end">
                <Button type="submit" :disabled="saving">
                  {{ t('common.save') }}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <!-- Tokens Management -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between">
          <CardTitle>{{ t('tokens.title') }}</CardTitle>
          <Button size="sm" class="gap-1" @click="createTokenOpen = true; newToken = null">
            <Plus class="h-4 w-4" />
            {{ t('tokens.create') }}
          </Button>
        </CardHeader>
        <CardContent>
          <Table v-if="tokens.length > 0">
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('tokens.publicToken') }}</TableHead>
                <TableHead>{{ t('tokens.name') }}</TableHead>
                <TableHead>{{ t('tokens.status') }}</TableHead>
                <TableHead>{{ t('tokens.createdAt') }}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="token in tokens" :key="token.id">
                <TableCell class="font-mono text-xs">
                  {{ token.public_token }}
                </TableCell>
                <TableCell>{{ token.name || '-' }}</TableCell>
                <TableCell>
                  <Badge v-if="token.is_active" variant="default" class="gap-1">
                    <ShieldCheck class="h-3 w-3" />
                    Active
                  </Badge>
                  <Badge v-else variant="secondary" class="gap-1">
                    <ShieldOff class="h-3 w-3" />
                    Revoked
                  </Badge>
                </TableCell>
                <TableCell>{{ new Date(token.created_at).toLocaleDateString() }}</TableCell>
                <TableCell>
                  <Button
                    v-if="token.is_active"
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 text-destructive"
                    @click="onConfirmRevoke(token)"
                  >
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <p v-else class="text-sm text-muted-foreground py-4 text-center">
            {{ t('tokens.noTokens') }}
          </p>
        </CardContent>
      </Card>
    </template>

    <div v-else class="py-12 text-center text-muted-foreground">
      {{ t('error.notFound') }}
    </div>

    <!-- Create Token Dialog -->
    <Dialog v-model:open="createTokenOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('tokens.create') }}</DialogTitle>
          <DialogDescription>{{ t('tokens.createDescription') }}</DialogDescription>
        </DialogHeader>

        <!-- Show new token (one-time) -->
        <template v-if="newToken">
          <div class="space-y-3">
            <div>
              <label class="text-sm font-medium">{{ t('tokens.publicToken') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <code class="flex-1 rounded bg-muted px-3 py-2 text-sm">{{ newToken.public_token }}</code>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-destructive">{{ t('tokens.privateToken') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <code class="flex-1 rounded bg-muted px-3 py-2 text-sm break-all">{{ newToken.private_token }}</code>
                <Button variant="outline" size="icon" @click="copyToClipboard(newToken!.private_token)">
                  <Copy class="h-4 w-4" />
                </Button>
              </div>
              <p class="mt-1 text-xs text-destructive">{{ t('tokens.privateWarning') }}</p>
            </div>
          </div>
          <DialogFooter>
            <Button @click="createTokenOpen = false">{{ t('common.close') }}</Button>
          </DialogFooter>
        </template>

        <!-- Create form -->
        <template v-else>
          <div class="space-y-4">
            <div>
              <label class="text-sm font-medium">{{ t('tokens.name') }}</label>
              <Input
                v-model="tokenName"
                :placeholder="t('tokens.namePlaceholder')"
                class="mt-1"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" @click="createTokenOpen = false">{{ t('common.cancel') }}</Button>
            <Button :disabled="creatingToken" @click="onCreateToken">{{ t('tokens.create') }}</Button>
          </DialogFooter>
        </template>
      </DialogContent>
    </Dialog>

    <!-- Revoke Confirmation -->
    <Dialog v-model:open="revokeOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('tokens.revokeTitle') }}</DialogTitle>
          <DialogDescription>{{ t('tokens.revokeDescription') }}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="revokeOpen = false">{{ t('common.cancel') }}</Button>
          <Button variant="destructive" @click="onRevoke">{{ t('tokens.revoke') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ArrowLeft, X } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { User } from '~/composables/useUsers'
import type { Role } from '~/composables/useRoles'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'
import { Separator } from '~/components/ui/separator'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))
const userId = computed(() => Number(route.params.userId))

const { fetchUsers, updateUser } = useUsers()
const { fetchRoles, assignRole, removeRole } = useRoles()

const user = ref<User | null>(null)
const allRoles = ref<Role[]>([])
const loading = ref(true)
const saving = ref(false)
const assigningRole = ref(false)
const selectedRoleId = ref<string>('')

const userSchema = toTypedSchema(
  z.object({
    last_name: z.string().min(1, t('users.validation.nameRequired')),
    first_name: z.string().min(1, t('users.validation.nameRequired')),
    email: z.string().email(t('users.validation.emailInvalid')),
  }),
)

const initialValues = computed(() => ({
  last_name: user.value?.last_name ?? '',
  first_name: user.value?.first_name ?? '',
  email: user.value?.email ?? '',
}))

const userRoles = computed(() => user.value?.roles ?? [])

const availableRoles = computed(() =>
  allRoles.value.filter(r => !userRoles.value.some(ur => ur.id === r.id)),
)

async function loadData() {
  loading.value = true
  try {
    const [usersResp, rolesResp] = await Promise.all([
      fetchUsers(orgId.value, 1, 100),
      fetchRoles(orgId.value),
    ])
    user.value = usersResp.data.find(u => u.id === userId.value) ?? null
    allRoles.value = rolesResp
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
    await updateUser(userId.value, {
      last_name: values.last_name as string,
      first_name: values.first_name as string,
      email: values.email as string,
    })
    toast({ title: t('users.updated') })
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

async function onAssignRole() {
  if (!selectedRoleId.value) return
  assigningRole.value = true
  try {
    await assignRole(Number(selectedRoleId.value), userId.value)
    selectedRoleId.value = ''
    toast({ title: t('users.roleAssigned') })
    await loadData()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    assigningRole.value = false
  }
}

async function onRemoveRole(roleId: number) {
  try {
    await removeRole(roleId, userId.value)
    toast({ title: t('users.roleRemoved') })
    await loadData()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
}

function goBack() {
  router.push(`/org/${orgId.value}/settings/users`)
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6">
    <Button variant="ghost" size="sm" class="gap-2" @click="goBack">
      <ArrowLeft class="h-4 w-4" />
      {{ t('users.backToUsers') }}
    </Button>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <span class="text-muted-foreground">{{ t('common.loading') }}</span>
    </div>

    <template v-else-if="user">
      <!-- User Details -->
      <Card>
        <CardHeader>
          <CardTitle>{{ t('users.editUser') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <Form
            v-slot="{ handleSubmit }"
            :validation-schema="userSchema"
            :initial-values="initialValues"
            class="space-y-4"
          >
            <form @submit="handleSubmit($event, onSave)">
              <div class="grid grid-cols-2 gap-4">
                <FormField v-slot="{ componentField }" name="last_name">
                  <FormItem>
                    <FormLabel>{{ t('users.lastName') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>

                <FormField v-slot="{ componentField }" name="first_name">
                  <FormItem>
                    <FormLabel>{{ t('users.firstName') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>

              <FormField v-slot="{ componentField }" name="email">
                <FormItem>
                  <FormLabel>{{ t('users.email') }}</FormLabel>
                  <FormControl>
                    <Input type="email" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <div class="mt-4 flex justify-end">
                <Button type="submit" :disabled="saving">
                  {{ t('common.save') }}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <!-- Roles Management -->
      <Card>
        <CardHeader>
          <CardTitle>{{ t('users.roles') }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <!-- Current Roles -->
          <div v-if="userRoles.length > 0" class="flex flex-wrap gap-2">
            <Badge
              v-for="role in userRoles"
              :key="role.id"
              variant="secondary"
              class="gap-1 pr-1"
            >
              {{ role.name }}
              <button
                class="ml-1 rounded-full p-0.5 hover:bg-muted"
                @click="onRemoveRole(role.id)"
              >
                <X class="h-3 w-3" />
              </button>
            </Badge>
          </div>
          <p v-else class="text-sm text-muted-foreground">
            {{ t('users.noRoles') }}
          </p>

          <Separator />

          <!-- Assign New Role -->
          <div class="flex items-end gap-2">
            <div class="flex-1">
              <label class="text-sm font-medium">{{ t('users.assignRole') }}</label>
              <Select v-model="selectedRoleId">
                <SelectTrigger class="mt-1">
                  <SelectValue :placeholder="t('users.selectRole')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="role in availableRoles"
                    :key="role.id"
                    :value="String(role.id)"
                  >
                    {{ role.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Button
              :disabled="!selectedRoleId || assigningRole"
              @click="onAssignRole"
            >
              {{ t('users.assign') }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </template>

    <div v-else class="py-12 text-center text-muted-foreground">
      {{ t('error.notFound') }}
    </div>
  </div>
</template>

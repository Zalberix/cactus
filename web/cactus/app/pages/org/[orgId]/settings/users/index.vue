<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Users, MoreHorizontal, Pencil, Trash2 } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { User } from '~/composables/useUsers'
import type { Role } from '~/composables/useRoles'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
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
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))

const { fetchUsers, createUser, deleteUser } = useUsers()
const { fetchRoles } = useRoles()

const users = ref<User[]>([])
const loading = ref(true)
const page = ref(1)
const pageCount = ref(1)
const pageSize = ref(20)
const search = ref('')
const roles = ref<Role[]>([])

const inviteOpen = ref(false)
const deleteOpen = ref(false)
const userToDelete = ref<User | null>(null)
const submitting = ref(false)

let searchTimeout: ReturnType<typeof setTimeout> | null = null

const inviteSchema = toTypedSchema(
  z.object({
    email: z.string().email(t('users.validation.emailInvalid')),
    last_name: z.string().min(1, t('users.validation.nameRequired')),
    first_name: z.string().min(1, t('users.validation.nameRequired')),
    password: z.string().min(6, t('users.validation.passwordMin')),
    role_id: z.string().optional(),
  }),
)

const columns: ColumnDef<User>[] = [
  {
    id: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('users.name') }),
    cell: ({ row }) => h('span', { class: 'font-medium' }, `${row.original.last_name} ${row.original.first_name}`),
  },
  {
    accessorKey: 'email',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('users.email') }),
  },
  {
    id: 'roles',
    header: t('users.role'),
    cell: ({ row }) => {
      const userRoles = row.original.roles ?? []
      if (userRoles.length === 0) {
        return h('span', { class: 'text-muted-foreground text-sm' }, '-')
      }
      return h('div', { class: 'flex gap-1 flex-wrap' },
        userRoles.map(r => h(Badge, { key: r.id, variant: 'secondary' }, () => r.name)),
      )
    },
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('users.joinedAt') }),
    cell: ({ row }) => {
      const date = new Date(row.getValue('created_at') as string)
      return h('span', { class: 'text-muted-foreground' }, date.toLocaleDateString())
    },
  },
  {
    id: 'actions',
    header: '',
    size: 50,
    cell: ({ row }) => {
      return h(DropdownMenu, {}, {
        default: () => [
          h(DropdownMenuTrigger, { asChild: true }, () =>
            h(Button, { variant: 'ghost', size: 'icon', class: 'h-8 w-8' }, () =>
              h(MoreHorizontal, { class: 'h-4 w-4' }),
            ),
          ),
          h(DropdownMenuContent, { align: 'end' }, () => [
            h(DropdownMenuItem, {
              onClick: () => onEditUser(row.original),
            }, () => [h(Pencil, { class: 'mr-2 h-4 w-4' }), t('common.edit')]),
            h(DropdownMenuItem, {
              class: 'text-destructive focus:text-destructive',
              onClick: () => onConfirmDelete(row.original),
            }, () => [h(Trash2, { class: 'mr-2 h-4 w-4' }), t('destructive.removeUser.confirm')]),
          ]),
        ],
      })
    },
  },
]

async function loadUsers() {
  loading.value = true
  try {
    const result = await fetchUsers(orgId.value, page.value, pageSize.value)
    users.value = result.data
    pageCount.value = result.meta.total_pages
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function loadRoles() {
  try {
    roles.value = await fetchRoles(orgId.value)
  }
  catch {
    // Non-critical, role dropdown will be empty
  }
}

function onSearch(value: string) {
  search.value = value
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    loadUsers()
  }, 300)
}

function onEditUser(u: User) {
  router.push(`/org/${orgId.value}/settings/users/${u.id}`)
}

function onConfirmDelete(user: User) {
  userToDelete.value = user
  deleteOpen.value = true
}

async function onDelete() {
  if (!userToDelete.value) return
  submitting.value = true
  try {
    await deleteUser(userToDelete.value.id)
    deleteOpen.value = false
    userToDelete.value = null
    toast({ title: t('users.deleted') })
    await loadUsers()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

async function onInvite(values: Record<string, unknown>) {
  submitting.value = true
  try {
    await createUser(orgId.value, {
      email: values.email as string,
      last_name: values.last_name as string,
      first_name: values.first_name as string,
      password: values.password as string,
      role_id: values.role_id ? Number(values.role_id) : undefined,
    })
    inviteOpen.value = false
    toast({ title: t('users.invited') })
    await loadUsers()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadUsers()
  loadRoles()
})

watch([page, pageSize], () => loadUsers())
</script>

<template>
  <div class="space-y-4">
    <!-- Toolbar -->
    <div class="flex items-center justify-between">
      <Input
        :model-value="search"
        :placeholder="t('common.search')"
        class="max-w-sm"
        @update:model-value="onSearch($event as string)"
      />
      <Button @click="inviteOpen = true">
        {{ t('users.invite') }}
      </Button>
    </div>

    <!-- Data Table -->
    <DataTable
      v-if="!loading || users.length > 0"
      :columns="columns"
      :data="users"
      :loading="loading"
      v-model:page="page"
      v-model:page-size="pageSize"
      :page-count="pageCount"
    >
      <template #empty>
        <EmptyState
          :icon="Users"
          :heading="t('empty.users.heading')"
          :body="t('empty.users.body')"
          :cta-label="t('empty.users.cta')"
          @cta="inviteOpen = true"
        />
      </template>
    </DataTable>

    <!-- Loading state -->
    <DataTable
      v-if="loading && users.length === 0"
      :columns="columns"
      :data="[]"
      :loading="true"
    />

    <!-- Invite User Dialog -->
    <Dialog v-model:open="inviteOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('users.invite') }}</DialogTitle>
          <DialogDescription>{{ t('users.inviteDescription') }}</DialogDescription>
        </DialogHeader>
        <Form
          v-slot="{ handleSubmit }"
          :validation-schema="inviteSchema"
          class="space-y-4"
        >
          <form @submit="handleSubmit($event, onInvite)">
            <div class="space-y-4">
              <FormField v-slot="{ componentField }" name="email">
                <FormItem>
                  <FormLabel>{{ t('users.email') }}</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      :placeholder="t('auth.emailPlaceholder')"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

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

              <FormField v-slot="{ componentField }" name="password">
                <FormItem>
                  <FormLabel>{{ t('auth.password') }}</FormLabel>
                  <FormControl>
                    <Input
                      type="password"
                      :placeholder="t('users.passwordPlaceholder')"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="role_id">
                <FormItem>
                  <FormLabel>{{ t('users.role') }}</FormLabel>
                  <FormControl>
                    <Select v-bind="componentField">
                      <SelectTrigger>
                        <SelectValue :placeholder="t('users.selectRole')" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem
                          v-for="role in roles"
                          :key="role.id"
                          :value="String(role.id)"
                        >
                          {{ role.name }}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <DialogFooter class="mt-6">
              <Button type="button" variant="outline" @click="inviteOpen = false">
                {{ t('common.cancel') }}
              </Button>
              <Button type="submit" :disabled="submitting">
                {{ t('users.invite') }}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation Dialog -->
    <Dialog v-model:open="deleteOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('destructive.removeUser.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.removeUser.body', { email: userToDelete?.email ?? '' }) }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" @click="deleteOpen = false">
            {{ t('destructive.cancel') }}
          </Button>
          <Button
            variant="destructive"
            :disabled="submitting"
            @click="onDelete"
          >
            {{ t('destructive.removeUser.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

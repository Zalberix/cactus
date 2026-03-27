<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Shield, MoreHorizontal, Pencil, Trash2 } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
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
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))

const { fetchRoles, createRole, deleteRole } = useRoles()

const roles = ref<Role[]>([])
const loading = ref(true)
const createOpen = ref(false)
const deleteOpen = ref(false)
const roleToDelete = ref<Role | null>(null)
const submitting = ref(false)

const createSchema = toTypedSchema(
  z.object({
    name: z.string().min(2, t('roles.validation.nameMin')),
    description: z.string().optional(),
  }),
)

const columns: ColumnDef<Role>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('roles.name') }),
    cell: ({ row }) => h('span', { class: 'font-medium' }, row.getValue('name')),
  },
  {
    accessorKey: 'description',
    header: t('roles.description'),
    cell: ({ row }) => {
      const desc = row.getValue('description') as string | undefined
      return h('span', { class: 'text-muted-foreground' }, desc || '-')
    },
  },
  {
    id: 'permissionsCount',
    header: t('roles.permissions'),
    cell: ({ row }) => {
      const count = row.original.permissions?.length ?? 0
      return h(Badge, { variant: 'secondary' }, () => t('roles.permissionsCount', { count }))
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
              onClick: () => onEditRole(row.original),
            }, () => [h(Pencil, { class: 'mr-2 h-4 w-4' }), t('common.edit')]),
            h(DropdownMenuItem, {
              class: 'text-destructive focus:text-destructive',
              onClick: () => onConfirmDelete(row.original),
            }, () => [h(Trash2, { class: 'mr-2 h-4 w-4' }), t('common.delete')]),
          ]),
        ],
      })
    },
  },
]

async function loadRoles() {
  loading.value = true
  try {
    roles.value = await fetchRoles(orgId.value)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function onEditRole(role: Role) {
  router.push(`/org/${orgId.value}/settings/roles/${role.id}`)
}

function onConfirmDelete(role: Role) {
  roleToDelete.value = role
  deleteOpen.value = true
}

async function onDelete() {
  if (!roleToDelete.value) return
  submitting.value = true
  try {
    await deleteRole(roleToDelete.value.id)
    deleteOpen.value = false
    roleToDelete.value = null
    toast({ title: t('roles.deleted') })
    await loadRoles()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

async function onCreate(values: Record<string, unknown>) {
  submitting.value = true
  try {
    const role = await createRole(orgId.value, {
      name: values.name as string,
      description: (values.description as string) || undefined,
      permissions: [],
    })
    createOpen.value = false
    toast({ title: t('roles.created') })
    // Navigate to role detail to set permissions
    router.push(`/org/${orgId.value}/settings/roles/${role.id}`)
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

function onRowClick(role: Role) {
  router.push(`/org/${orgId.value}/settings/roles/${role.id}`)
}

onMounted(() => {
  loadRoles()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Toolbar -->
    <div class="flex items-center justify-end">
      <Button @click="createOpen = true">
        {{ t('roles.create') }}
      </Button>
    </div>

    <!-- Data Table -->
    <DataTable
      v-if="!loading || roles.length > 0"
      :columns="columns"
      :data="roles"
      :loading="loading"
    >
      <template #empty>
        <EmptyState
          :icon="Shield"
          :heading="t('empty.roles.heading')"
          :body="t('empty.roles.body')"
          :cta-label="t('empty.roles.cta')"
          @cta="createOpen = true"
        />
      </template>
    </DataTable>

    <DataTable
      v-if="loading && roles.length === 0"
      :columns="columns"
      :data="[]"
      :loading="true"
    />

    <!-- Create Role Dialog -->
    <Dialog v-model:open="createOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('roles.create') }}</DialogTitle>
          <DialogDescription>{{ t('roles.createDescription') }}</DialogDescription>
        </DialogHeader>
        <Form
          v-slot="{ handleSubmit }"
          :validation-schema="createSchema"
          class="space-y-4"
        >
          <form @submit="handleSubmit($event, onCreate)">
            <div class="space-y-4">
              <FormField v-slot="{ componentField }" name="name">
                <FormItem>
                  <FormLabel>{{ t('roles.name') }}</FormLabel>
                  <FormControl>
                    <Input
                      :placeholder="t('roles.namePlaceholder')"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="description">
                <FormItem>
                  <FormLabel>{{ t('roles.description') }}</FormLabel>
                  <FormControl>
                    <Input
                      :placeholder="t('roles.descriptionPlaceholder')"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <DialogFooter class="mt-6">
              <Button type="button" variant="outline" @click="createOpen = false">
                {{ t('common.cancel') }}
              </Button>
              <Button type="submit" :disabled="submitting">
                {{ t('roles.create') }}
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
          <DialogTitle>{{ t('destructive.deleteRole.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deleteRole.body', { name: roleToDelete?.name ?? '' }) }}
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
            {{ t('destructive.deleteRole.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Box, MoreHorizontal, Pencil, Trash2, Key } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { System } from '~/composables/useSystems'
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

const { fetchSystemsPage, createSystem, deleteSystem } = useSystems()

const systems = ref<System[]>([])
const loading = ref(true)
const page = ref(1)
const pageCount = ref(1)
const pageSize = ref(20)
const createOpen = ref(false)
const deleteOpen = ref(false)
const systemToDelete = ref<System | null>(null)
const submitting = ref(false)

const createSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t('systems.validation.nameRequired')),
    description: z.string().optional(),
  }),
)

const columns: ColumnDef<System>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('systems.name') }),
    cell: ({ row }) => h('span', { class: 'font-medium' }, row.getValue('name')),
  },
  {
    accessorKey: 'description',
    header: t('systems.description'),
    cell: ({ row }) => {
      const desc = row.getValue('description') as string | undefined
      return h('span', { class: 'text-muted-foreground' }, desc || '-')
    },
  },
  {
    id: 'tokensCount',
    header: t('systems.tokensCount'),
    cell: ({ row }) => {
      const count = row.original.active_tokens_count ?? 0
      return h(Badge, { variant: 'secondary' }, () => String(count))
    },
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('systems.createdAt') }),
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
              onClick: () => onManageTokens(row.original),
            }, () => [h(Key, { class: 'mr-2 h-4 w-4' }), t('systems.manageTokens')]),
            h(DropdownMenuItem, {
              onClick: () => onEditSystem(row.original),
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

async function loadSystems() {
  loading.value = true
  try {
    const result = await fetchSystemsPage(orgId.value, page.value, pageSize.value)
    systems.value = result.data
    pageCount.value = result.meta.total_pages
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function onManageTokens(system: System) {
  router.push(`/org/${orgId.value}/settings/systems/${system.id}/tokens`)
}

function onEditSystem(s: System) {
  router.push(`/org/${orgId.value}/settings/systems/${s.id}`)
}

function onConfirmDelete(system: System) {
  systemToDelete.value = system
  deleteOpen.value = true
}

async function onDelete() {
  if (!systemToDelete.value) return
  submitting.value = true
  try {
    await deleteSystem(systemToDelete.value.id)
    deleteOpen.value = false
    systemToDelete.value = null
    toast({ title: t('systems.deleted') })
    await loadSystems()
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
    await createSystem(
      orgId.value,
      values.name as string,
      (values.description as string) || undefined,
    )
    createOpen.value = false
    toast({ title: t('systems.created') })
    await loadSystems()
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadSystems()
})

watch([page, pageSize], () => loadSystems())
</script>

<template>
  <div class="space-y-4">
    <!-- Toolbar -->
    <div class="flex items-center justify-end">
      <Button @click="createOpen = true">
        {{ t('systems.create') }}
      </Button>
    </div>

    <!-- Data Table -->
    <DataTable
      v-if="!loading || systems.length > 0"
      :columns="columns"
      :data="systems"
      :loading="loading"
      v-model:page="page"
      v-model:page-size="pageSize"
      :page-count="pageCount"
    >
      <template #empty>
        <EmptyState
          :icon="Box"
          :heading="t('empty.systems.heading')"
          :body="t('empty.systems.body')"
          :cta-label="t('empty.systems.cta')"
          @cta="createOpen = true"
        />
      </template>
    </DataTable>

    <DataTable
      v-if="loading && systems.length === 0"
      :columns="columns"
      :data="[]"
      :loading="true"
    />

    <!-- Create System Dialog -->
    <Dialog v-model:open="createOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('systems.create') }}</DialogTitle>
          <DialogDescription>{{ t('systems.createDescription') }}</DialogDescription>
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
                  <FormLabel>{{ t('systems.name') }}</FormLabel>
                  <FormControl>
                    <Input
                      :placeholder="t('systems.namePlaceholder')"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="description">
                <FormItem>
                  <FormLabel>{{ t('systems.description') }}</FormLabel>
                  <FormControl>
                    <Input
                      :placeholder="t('systems.descriptionPlaceholder')"
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
                {{ t('systems.create') }}
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
          <DialogTitle>{{ t('destructive.deleteSystem.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deleteSystem.body', { name: systemToDelete?.name ?? '' }) }}
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
            {{ t('destructive.deleteSystem.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

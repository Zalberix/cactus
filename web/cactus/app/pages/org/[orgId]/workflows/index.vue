<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'
import { h } from 'vue'
import { Workflow as WorkflowIcon, MoreHorizontal, Pencil, Trash2, Search } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { Workflow } from '~/composables/useWorkflows'
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

const { fetchWorkflowsForOrg, createWorkflow, deleteWorkflow } = useWorkflows()
const { fetchSystems } = useSystems()

const workflows = ref<Workflow[]>([])
const systems = ref<System[]>([])
const loading = ref(true)
const createOpen = ref(false)
const deleteOpen = ref(false)
const workflowToDelete = ref<Workflow | null>(null)
const submitting = ref(false)
const searchQuery = ref('')

const priorityLabels: Record<number, string> = {
  0: 'Low',
  1: 'Normal',
  2: 'High',
  3: 'Urgent',
}

const priorityColors: Record<number, string> = {
  0: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  1: 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300',
  2: 'bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300',
  3: 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300',
}

const createSchema = toTypedSchema(
  z.object({
    system_id: z.string().min(1, 'System is required'),
    name: z.string().min(3, 'Name must be at least 3 characters'),
    priority: z.string().default('1'),
  }),
)

let searchTimeout: ReturnType<typeof setTimeout> | null = null
const debouncedSearch = ref('')

watch(searchQuery, (val) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    debouncedSearch.value = val
  }, 300)
})

const filteredWorkflows = computed(() => {
  if (!debouncedSearch.value) return workflows.value
  const q = debouncedSearch.value.toLowerCase()
  return workflows.value.filter(wf =>
    wf.name.toLowerCase().includes(q),
  )
})

const columns: ColumnDef<Workflow>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workflows.name') }),
    cell: ({ row }) => h(
      resolveComponent('NuxtLink') as any,
      {
        to: `/org/${orgId.value}/workflows/${row.original.id}/edit`,
        class: 'font-medium text-primary hover:underline',
      },
      () => row.getValue('name'),
    ),
  },
  {
    accessorKey: 'system_name',
    header: t('workflows.system'),
    cell: ({ row }) => {
      const name = row.getValue('system_name') as string | undefined
      return name
        ? h(Badge, { variant: 'secondary' }, () => name)
        : h('span', { class: 'text-muted-foreground' }, '-')
    },
  },
  {
    accessorKey: 'priority',
    header: t('workflows.priority'),
    cell: ({ row }) => {
      const priority = row.getValue('priority') as number
      const label = priorityLabels[priority] ?? 'Unknown'
      const colorClass = priorityColors[priority] ?? priorityColors[1]
      return h(
        'span',
        {
          class: `inline-flex items-center rounded-full border border-transparent px-2.5 py-0.5 text-xs font-semibold ${colorClass}`,
        },
        label,
      )
    },
  },
  {
    id: 'activeVersions',
    header: t('editor.versions'),
    cell: ({ row }) => {
      const count = row.original.active_version_count ?? 0
      return h(Badge, { variant: 'secondary' }, () => String(count))
    },
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => h(DataTableColumnHeader, { column: column as any, title: t('workflows.createdAt') }),
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
              onClick: () => onEditWorkflow(row.original),
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

async function loadData() {
  loading.value = true
  try {
    const [wfs, sys] = await Promise.all([
      fetchWorkflowsForOrg(orgId.value),
      fetchSystems(orgId.value),
    ])
    workflows.value = wfs
    systems.value = sys
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function onEditWorkflow(workflow: Workflow) {
  router.push(`/org/${orgId.value}/workflows/${workflow.id}/edit`)
}

function onConfirmDelete(workflow: Workflow) {
  workflowToDelete.value = workflow
  deleteOpen.value = true
}

async function onDelete() {
  if (!workflowToDelete.value) return
  submitting.value = true
  try {
    await deleteWorkflow(workflowToDelete.value.id)
    deleteOpen.value = false
    workflowToDelete.value = null
    toast({ title: t('workflows.deleted') })
    await loadData()
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    submitting.value = false
  }
}

async function onCreate(values: Record<string, unknown>) {
  submitting.value = true
  try {
    const wf = await createWorkflow(
      Number(values.system_id),
      {
        name: values.name as string,
        priority: Number(values.priority),
      },
    )
    createOpen.value = false
    toast({ title: t('workflows.created') })
    router.push(`/org/${orgId.value}/workflows/${wf.id}/edit`)
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
    <!-- Toolbar -->
    <div class="flex items-center justify-between gap-4">
      <div class="relative w-72">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          :placeholder="t('common.search')"
          class="pl-9"
        />
      </div>
      <Button @click="createOpen = true">
        {{ t('workflows.create') }}
      </Button>
    </div>

    <!-- Data Table -->
    <DataTable
      v-if="!loading || workflows.length > 0"
      :columns="columns"
      :data="filteredWorkflows"
      :loading="loading"
    >
      <template #empty>
        <EmptyState
          :icon="WorkflowIcon"
          :heading="t('empty.workflows.heading')"
          :body="t('empty.workflows.body')"
          :cta-label="t('empty.workflows.cta')"
          @cta="createOpen = true"
        />
      </template>
    </DataTable>

    <DataTable
      v-if="loading && workflows.length === 0"
      :columns="columns"
      :data="[]"
      :loading="true"
    />

    <!-- Create Workflow Dialog -->
    <Dialog v-model:open="createOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('workflows.create') }}</DialogTitle>
          <DialogDescription>
            Create a new workflow to orchestrate notification delivery.
          </DialogDescription>
        </DialogHeader>
        <Form
          v-slot="{ handleSubmit }"
          :validation-schema="createSchema"
          class="space-y-4"
        >
          <form @submit="handleSubmit($event, onCreate)">
            <div class="space-y-4">
              <FormField v-slot="{ componentField }" name="system_id">
                <FormItem>
                  <FormLabel>{{ t('workflows.system') }}</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a system" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem
                        v-for="sys in systems"
                        :key="sys.id"
                        :value="String(sys.id)"
                      >
                        {{ sys.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="name">
                <FormItem>
                  <FormLabel>{{ t('workflows.name') }}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="e.g. Welcome Email Flow"
                      v-bind="componentField"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="priority">
                <FormItem>
                  <FormLabel>{{ t('workflows.priority') }}</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select priority" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="0">Low</SelectItem>
                      <SelectItem value="1">Normal</SelectItem>
                      <SelectItem value="2">High</SelectItem>
                      <SelectItem value="3">Urgent</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <DialogFooter class="mt-6">
              <Button type="button" variant="outline" @click="createOpen = false">
                {{ t('common.cancel') }}
              </Button>
              <Button type="submit" :disabled="submitting">
                {{ t('workflows.create') }}
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
          <DialogTitle>{{ t('destructive.deleteWorkflow.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('destructive.deleteWorkflow.body', { name: workflowToDelete?.name ?? '' }) }}
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
            {{ t('destructive.deleteWorkflow.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

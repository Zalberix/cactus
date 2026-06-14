<script setup lang="ts">
import type { RoutingVersionRowRecord } from '~/composables/useWorkflowRouting'
import type { ColumnDef, PaginationState } from '@tanstack/vue-table'
import {
  getCoreRowModel,
  useVueTable,
} from '@tanstack/vue-table'
import { MoreHorizontal } from 'lucide-vue-next'
import DataTablePagination from '~/components/tables/DataTablePagination.vue'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'

const props = withDefaults(defineProps<{
  rows: RoutingVersionRowRecord[]
  saving?: boolean
  page?: number
  pageCount?: number
  pageSize?: number
  editPath?: (row: RoutingVersionRowRecord) => string
}>(), {
  page: 1,
  pageCount: undefined,
  pageSize: 20,
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
  edit: [row: RoutingVersionRowRecord]
  compatibility: [row: RoutingVersionRowRecord]
  archive: [row: RoutingVersionRowRecord]
}>()

const { t } = useI18n()
const columns: ColumnDef<RoutingVersionRowRecord>[] = [
  { id: 'schema' },
  { id: 'supportedVersions' },
  { id: 'activeTesting' },
  { id: 'actions' },
]

const pagination = ref<PaginationState>({
  pageIndex: props.page - 1,
  pageSize: props.pageSize,
})

watch(
  () => [props.page, props.pageSize] as const,
  ([page, pageSize]) => {
    pagination.value = {
      pageIndex: page - 1,
      pageSize,
    }
  },
)

const table = useVueTable({
  get data() { return props.rows },
  get columns() { return columns },
  get pageCount() { return props.pageCount },
  get manualPagination() { return props.pageCount !== undefined },
  getCoreRowModel: getCoreRowModel(),
  onPaginationChange: (updaterOrValue) => {
    const next = typeof updaterOrValue === 'function'
      ? updaterOrValue(pagination.value)
      : updaterOrValue

    const normalized = {
      ...next,
      pageIndex: next.pageSize !== pagination.value.pageSize ? 0 : next.pageIndex,
    }

    pagination.value = normalized
    emit('update:page', normalized.pageIndex + 1)
    emit('update:pageSize', normalized.pageSize)
  },
  state: {
    get pagination() { return pagination.value },
  },
})

function schemaVersionLabel(row: RoutingVersionRowRecord) {
  const code = row.input_schema_code.trim()
  const versionNumber = String(row.input_schema_version_number)
  if (!code) {
    return `v${versionNumber}`
  }
  return code.includes(versionNumber) ? code : `${code} v${versionNumber}`
}

function isNativeSupport(version: RoutingVersionRowRecord['supported_versions'][number]) {
  return version.support_mode === 'native'
    && version.compatibility_type === 'native'
    && version.workflow_input_mapper_id === undefined
}

function processVersionLabel(version: RoutingVersionRowRecord['supported_versions'][number]) {
  const name = version.workflow_version_name.trim()
  const number = String(version.workflow_version_number)
  return !name
    ? `v${number}`
    : name.includes(number) ? name : `${name} v${number}`
}

function versionLabel(version: RoutingVersionRowRecord['supported_versions'][number]) {
  const processVersion = processVersionLabel(version)
  if (isNativeSupport(version)) {
    return processVersion
  }

  const support = version.workflow_input_mapper_id
    ? `Mapper #${version.workflow_input_mapper_id}`
    : version.compatibility_type && version.compatibility_type !== 'native'
      ? version.compatibility_type
      : t('workflowRouting.compatibility')
  return `${processVersion} - ${support}`
}

function statusLabel(status: RoutingVersionRowRecord['input_schema_status']) {
  switch (status) {
    case 'active':
      return t('workflowRouting.statusActive')
    case 'deprecated':
      return t('workflowRouting.statusDeprecated')
    case 'archived':
      return t('workflowRouting.statusArchived')
    default:
      return t('workflowRouting.statusDraft')
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="overflow-hidden rounded-2xl border bg-background">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{{ t('workflowRouting.schemaVersion') }}</TableHead>
            <TableHead>{{ t('workflowRouting.supportedProcessVersions') }}</TableHead>
            <TableHead>{{ t('workflowRouting.activeTesting') }}</TableHead>
            <TableHead class="w-12 text-right">{{ t('common.actions') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="rows.length === 0" :colspan="4">
            {{ t('workflowRouting.noRoutingVersions') }}
          </TableEmpty>
        <TableRow v-for="row in rows" :key="row.input_schema_id" :data-testid="`routing-schema-row-${row.input_schema_id}`">
            <TableCell>
              <div class="flex items-center gap-2 whitespace-nowrap" :data-testid="`routing-schema-inline-${row.input_schema_id}`">
                <NuxtLink
                  v-if="editPath"
                  :to="editPath(row)"
                  class="font-medium hover:underline"
                >
                  {{ schemaVersionLabel(row) }}
                </NuxtLink>
                <button
                  v-else
                  type="button"
                  class="font-medium underline hover:text-foreground/80"
                  @click="emit('edit', row)"
                >
                  {{ schemaVersionLabel(row) }}
                </button>
                <Badge variant="secondary">{{ statusLabel(row.input_schema_status) }}</Badge>
                <Badge v-if="row.is_default" variant="outline">{{ t('workflowRouting.fieldDefaultSchema') }}</Badge>
              </div>
            </TableCell>
            <TableCell>
              <div v-if="row.supported_versions.length > 0" class="flex flex-wrap gap-1.5">
                <Badge
                  v-for="version in row.supported_versions"
                  :key="version.compatibility_id"
                  variant="outline"
                  :class="isNativeSupport(version) ? 'border-zinc-900 bg-zinc-900 text-zinc-50 hover:bg-zinc-800' : undefined"
                >
                  {{ versionLabel(version) }}
                </Badge>
              </div>
              <span v-else class="text-sm text-muted-foreground">-</span>
            </TableCell>
            <TableCell>
              <div v-if="row.active_tests.length > 0" class="flex flex-wrap gap-1.5">
                <Badge v-for="test in row.active_tests" :key="test.id" variant="secondary">
                  {{ test.name }}
                </Badge>
              </div>
              <span v-else class="text-sm text-muted-foreground">-</span>
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" class="h-8 w-8" :disabled="saving" :aria-label="t('common.actions')">
                    <MoreHorizontal class="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem @click="emit('edit', row)">
                    {{ t('common.edit') }}
                  </DropdownMenuItem>
                  <DropdownMenuItem @click="emit('compatibility', row)">
                    {{ t('workflowRouting.compatibility') }}
                  </DropdownMenuItem>
                  <DropdownMenuItem class="text-destructive focus:text-destructive" @click="emit('archive', row)">
                    {{ t('common.delete') }}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <DataTablePagination
      v-if="rows.length > 0"
      :table="(table as any)"
    />
  </div>
</template>

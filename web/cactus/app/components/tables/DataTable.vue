<script setup lang="ts" generic="TData, TValue">
import {
  FlexRender,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  useVueTable,
} from '@tanstack/vue-table'
import type { ColumnDef, PaginationState, SortingState } from '@tanstack/vue-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import { Skeleton } from '~/components/ui/skeleton'
import DataTablePagination from './DataTablePagination.vue'

const props = withDefaults(defineProps<{
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  loading?: boolean
  page?: number
  pageCount?: number
  pageSize?: number
}>(), {
  loading: false,
  page: 1,
  pageCount: undefined,
  pageSize: 20,
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
}>()

const sorting = ref<SortingState>([])
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
  get data() { return props.data },
  get columns() { return props.columns },
  get pageCount() { return props.pageCount },
  get manualPagination() { return props.pageCount !== undefined },
  getCoreRowModel: getCoreRowModel(),
  getSortedRowModel: getSortedRowModel(),
  getPaginationRowModel: getPaginationRowModel(),
  onSortingChange: (updaterOrValue) => {
    sorting.value = typeof updaterOrValue === 'function'
      ? updaterOrValue(sorting.value)
      : updaterOrValue
  },
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
    get sorting() { return sorting.value },
    get pagination() { return pagination.value },
  },
})

const skeletonRows = computed(() =>
  Array.from({ length: Math.min(props.pageSize, 5) }, (_, i) => i),
)
</script>

<template>
  <div class="space-y-4">
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
          >
            <TableHead
              v-for="header in headerGroup.headers"
              :key="header.id"
              :style="{ width: header.getSize() !== 150 ? `${header.getSize()}px` : undefined }"
            >
              <FlexRender
                v-if="!header.isPlaceholder"
                :render="header.column.columnDef.header"
                :props="header.getContext()"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <!-- Loading skeleton rows -->
          <template v-if="loading">
            <TableRow v-for="row in skeletonRows" :key="`skeleton-${row}`">
              <TableCell
                v-for="col in columns.length"
                :key="`skeleton-${row}-${col}`"
              >
                <Skeleton class="h-5 w-full" />
              </TableCell>
            </TableRow>
          </template>

          <!-- Data rows -->
          <template v-else-if="table.getRowModel().rows.length">
            <TableRow
              v-for="row in table.getRowModel().rows"
              :key="row.id"
              :data-state="row.getIsSelected() ? 'selected' : undefined"
            >
              <TableCell
                v-for="cell in row.getVisibleCells()"
                :key="cell.id"
              >
                <FlexRender
                  :render="cell.column.columnDef.cell"
                  :props="cell.getContext()"
                />
              </TableCell>
            </TableRow>
          </template>

          <!-- Empty state -->
          <template v-else>
            <TableRow>
              <TableCell :colspan="columns.length" class="h-24 text-center">
                <slot name="empty">
                  {{ $t('common.noResults') }}
                </slot>
              </TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>

    <DataTablePagination
      v-if="!loading && table.getRowModel().rows.length > 0"
      :table="(table as any)"
    />
  </div>
</template>

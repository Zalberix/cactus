<script setup lang="ts">
import type { Table } from '@tanstack/vue-table'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-vue-next'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

const props = defineProps<{
  table: Table<unknown>
  pageSizes?: number[]
}>()

const sizes = computed(() => props.pageSizes ?? [10, 20, 50, 100])

const { t } = useI18n()
</script>

<template>
  <div class="flex items-center justify-between px-2 py-4">
    <div class="flex items-center space-x-2">
      <p class="text-sm font-medium">
        {{ t('common.rowsPerPage') }}
      </p>
      <Select
        :model-value="String(table.getState().pagination.pageSize)"
        @update:model-value="(val: string) => table.setPageSize(Number(val))"
      >
        <SelectTrigger class="h-8 w-[70px]">
          <SelectValue :placeholder="String(table.getState().pagination.pageSize)" />
        </SelectTrigger>
        <SelectContent side="top">
          <SelectItem
            v-for="size in sizes"
            :key="size"
            :value="String(size)"
          >
            {{ size }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="flex items-center space-x-6 lg:space-x-8">
      <div class="flex w-[120px] items-center justify-center text-sm font-medium">
        {{ t('common.page', { page: table.getState().pagination.pageIndex + 1, total: table.getPageCount() }) }}
      </div>
      <div class="flex items-center space-x-2">
        <Button
          variant="outline"
          size="icon"
          class="hidden h-8 w-8 lg:flex"
          :disabled="!table.getCanPreviousPage()"
          @click="table.setPageIndex(0)"
        >
          <ChevronsLeft class="size-4" />
          <span class="sr-only">First page</span>
        </Button>
        <Button
          variant="outline"
          size="icon"
          class="h-8 w-8"
          :disabled="!table.getCanPreviousPage()"
          @click="table.previousPage()"
        >
          <ChevronLeft class="size-4" />
          <span class="sr-only">{{ t('common.previous') }}</span>
        </Button>
        <Button
          variant="outline"
          size="icon"
          class="h-8 w-8"
          :disabled="!table.getCanNextPage()"
          @click="table.nextPage()"
        >
          <ChevronRight class="size-4" />
          <span class="sr-only">{{ t('common.next') }}</span>
        </Button>
        <Button
          variant="outline"
          size="icon"
          class="hidden h-8 w-8 lg:flex"
          :disabled="!table.getCanNextPage()"
          @click="table.setPageIndex(table.getPageCount() - 1)"
        >
          <ChevronsRight class="size-4" />
          <span class="sr-only">Last page</span>
        </Button>
      </div>
    </div>
  </div>
</template>

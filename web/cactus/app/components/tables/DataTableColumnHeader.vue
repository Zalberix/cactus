<script setup lang="ts">
import type { Column } from '@tanstack/vue-table'
import { ArrowDown, ArrowUp, ChevronsUpDown } from 'lucide-vue-next'
import { cn } from '~/utils/cn'

const props = defineProps<{
  column: Column<unknown, unknown>
  title: string
}>()

function toggleSort() {
  props.column.toggleSorting()
}
</script>

<template>
  <div v-if="column.getCanSort()" class="flex items-center space-x-2">
    <Button
      variant="ghost"
      size="sm"
      class="-ml-3 h-8 data-[state=open]:bg-accent"
      @click="toggleSort"
    >
      <span>{{ title }}</span>
      <ArrowDown
        v-if="column.getIsSorted() === 'desc'"
        class="ml-2 size-4"
      />
      <ArrowUp
        v-else-if="column.getIsSorted() === 'asc'"
        class="ml-2 size-4"
      />
      <ChevronsUpDown
        v-else
        class="ml-2 size-4"
      />
    </Button>
  </div>
  <div v-else :class="cn('text-sm font-medium')">
    {{ title }}
  </div>
</template>

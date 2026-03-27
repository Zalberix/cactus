<script setup lang="ts">
import { cn } from '~/utils/cn'

type StatusType = 'pending' | 'running' | 'done' | 'error' | 'working' | 'ready' | 'offline'

const props = defineProps<{
  status: StatusType
}>()

const statusClasses: Record<StatusType, string> = {
  pending: 'bg-muted text-muted-foreground',
  running: 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300',
  done: 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300',
  ready: 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300',
  error: 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300',
  working: 'bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300',
  offline: 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400',
}

const statusLabel = computed(() => {
  return props.status.charAt(0).toUpperCase() + props.status.slice(1)
})

const badgeClass = computed(() => {
  return cn(
    'inline-flex items-center rounded-full border border-transparent px-2.5 py-0.5 text-xs font-semibold transition-colors',
    statusClasses[props.status] ?? statusClasses.pending,
  )
})
</script>

<template>
  <span :class="badgeClass">
    {{ statusLabel }}
  </span>
</template>

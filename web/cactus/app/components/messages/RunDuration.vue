<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  startedAt?: string
  completedAt?: string
  durationMs?: number
}>()

function formatDuration(ms: number): string {
  const diff = Math.max(0, ms)
  if (diff < 1000) return `${diff}ms`
  if (diff < 60000) return `${Math.floor(diff / 1000)}s`
  return `${Math.floor(diff / 60000)}m ${Math.floor((diff % 60000) / 1000)}s`
}

const label = computed(() => {
  return typeof props.durationMs === 'number' ? formatDuration(props.durationMs) : '-'
})
</script>

<template>
  <span class="font-mono text-sm">{{ label }}</span>
</template>

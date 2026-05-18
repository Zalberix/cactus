<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  startedAt?: string
  completedAt?: string
}>()

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

const isLive = computed(() => Boolean(props.startedAt && !props.completedAt))

function formatDuration(ms: number): string {
  const diff = Math.max(0, ms)
  if (diff < 1000) return `${diff}ms`
  if (diff < 60000) return `${Math.floor(diff / 1000)}s`
  return `${Math.floor(diff / 60000)}m ${Math.floor((diff % 60000) / 1000)}s`
}

const label = computed(() => {
  if (!props.startedAt) return '-'
  const start = new Date(props.startedAt).getTime()
  const end = props.completedAt ? new Date(props.completedAt).getTime() : now.value
  return formatDuration(end - start)
})

onMounted(() => {
  timer = setInterval(() => {
    if (isLive.value) now.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <span class="font-mono text-sm">{{ label }}</span>
</template>

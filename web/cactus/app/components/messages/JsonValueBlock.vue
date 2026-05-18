<script setup lang="ts">
import { cn } from '~/utils/cn'

const props = withDefaults(defineProps<{
  title: string
  value?: unknown
  stateText?: string
  tone?: 'default' | 'muted' | 'warning' | 'error'
}>(), {
  tone: 'default',
})

const toneClass = computed(() => {
  const map = {
    default: 'border-border bg-muted/30 text-foreground',
    muted: 'border-border bg-muted/30 text-muted-foreground',
    warning: 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-200',
    error: 'border-red-200 bg-red-50 text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200',
  }
  return map[props.tone]
})

const jsonText = computed(() => JSON.stringify(props.value, null, 2))
</script>

<template>
  <section class="space-y-2">
    <h3 class="text-xs font-semibold uppercase text-muted-foreground">
      {{ title }}
    </h3>
    <div
      v-if="value === undefined"
      :class="cn('rounded-md border px-3 py-2 text-sm', toneClass)"
    >
      {{ stateText ?? '-' }}
    </div>
    <pre
      v-else
      class="max-h-56 overflow-auto rounded-md border bg-muted/30 p-3 text-xs leading-relaxed"
    >{{ jsonText }}</pre>
  </section>
</template>

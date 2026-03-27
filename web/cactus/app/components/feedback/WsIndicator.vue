<script setup lang="ts">
const props = defineProps<{
  connected: boolean
}>()

const { t } = useI18n()
</script>

<template>
  <div class="flex items-center gap-1.5">
    <span
      :class="[
        'inline-block size-2 rounded-full',
        connected ? 'bg-green-500' : 'bg-red-500 animate-pulse',
      ]"
    />
    <span class="text-xs text-muted-foreground">
      {{ connected ? t('messages.detail.connected') : t('messages.detail.disconnected') }}
    </span>
  </div>
</template>

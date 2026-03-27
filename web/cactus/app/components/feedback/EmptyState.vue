<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon?: Component
  heading: string
  body: string
  ctaLabel?: string
  ctaTo?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <div v-if="icon" class="mb-4">
      <component
        :is="icon"
        class="size-12 text-muted-foreground"
      />
    </div>
    <h3 class="text-2xl font-semibold text-foreground">
      {{ heading }}
    </h3>
    <p class="mt-2 max-w-md text-sm text-muted-foreground">
      {{ body }}
    </p>
    <div v-if="ctaLabel" class="mt-6">
      <Button v-if="ctaTo" as-child>
        <NuxtLink :to="ctaTo">
          {{ ctaLabel }}
        </NuxtLink>
      </Button>
      <Button v-else @click="$emit('cta')">
        {{ ctaLabel }}
      </Button>
    </div>
  </div>
</template>

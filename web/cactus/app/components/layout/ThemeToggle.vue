<script setup lang="ts">
import { Moon, Sun } from 'lucide-vue-next'

const colorMode = useColorMode()

function toggle() {
  colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
}
</script>

<template>
  <Button
    variant="ghost"
    size="icon"
    class="h-8 w-8"
    @click="toggle"
  >
    <Sun v-if="colorMode.value === 'light'" class="size-4" />
    <Moon v-else class="size-4" />
    <span class="sr-only">{{ colorMode.value === 'dark' ? $t('theme.light') : $t('theme.dark') }}</span>
  </Button>
</template>

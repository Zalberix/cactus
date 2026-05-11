<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'

const props = defineProps<{
  schema: Record<string, unknown> | null
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const { t } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const schemaText = computed(() => JSON.stringify(props.schema ?? {}, null, 2))
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle>{{ t('editor.inputSchema') }}</DialogTitle>
        <DialogDescription>{{ t('editor.inputSchemaDescription') }}</DialogDescription>
      </DialogHeader>

      <pre class="max-h-[520px] overflow-auto rounded-md border bg-muted/40 p-4 text-xs"><code>{{ schemaText }}</code></pre>
    </DialogContent>
  </Dialog>
</template>

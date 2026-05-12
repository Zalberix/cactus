<script setup lang="ts">
import type { WorkerSettingsSchemaSummary } from '~/composables/useWorkers'
import { schemaChoiceOptions } from '~/components/dag/step-toolbar-utils'
import { Button } from '~/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '~/components/ui/dialog'

const props = defineProps<{
  open: boolean
  stepName?: string
  schemas: WorkerSettingsSchemaSummary[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  choose: [schemaId: number]
}>()

const { t, locale } = useI18n()

const isOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const choices = computed(() => schemaChoiceOptions(props.schemas))

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

function schemaText(schema: WorkerSettingsSchemaSummary) {
  return JSON.stringify(schema.settings_schema ?? {}, null, 2)
}
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="max-w-3xl">
      <DialogHeader>
        <DialogTitle>{{ t('toolbar.chooseSchema') }}</DialogTitle>
        <DialogDescription>{{ stepName }}</DialogDescription>
      </DialogHeader>

      <div class="max-h-[620px] space-y-3 overflow-auto pr-1">
        <div
          v-for="schema in choices"
          :key="schema.id"
          class="rounded-md border p-3"
        >
          <div class="flex flex-wrap items-start gap-3">
            <div class="min-w-0 flex-1 space-y-1">
              <div class="truncate text-sm font-medium">
                {{ t('toolbar.schemaVersion', { version: schema.version }) }}
              </div>
              <div class="text-xs text-muted-foreground">
                {{ t('toolbar.schemaAddedAt', { date: formatDate(schema.created_at) }) }}
              </div>
              <div class="text-xs text-muted-foreground">
                {{ t('toolbar.schemaWorkers', { ready: schema.ready_workers, total: schema.worker_count }) }}
              </div>
            </div>
            <Button
              type="button"
              size="sm"
              @click="emit('choose', schema.id)"
            >
              {{ t('toolbar.useSchema') }}
            </Button>
          </div>

          <details class="mt-3">
            <summary class="cursor-pointer text-xs font-medium text-muted-foreground">
              {{ t('toolbar.settingsSchemaPreview') }}
            </summary>
            <pre class="mt-2 max-h-72 overflow-auto rounded-md border bg-muted/40 p-3 text-xs"><code>{{ schemaText(schema) }}</code></pre>
          </details>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

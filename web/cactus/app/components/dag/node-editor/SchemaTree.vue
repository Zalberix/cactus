<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '~/components/ui/collapsible'

interface JsonSchemaProperty {
  type?: string
  properties?: Record<string, JsonSchemaProperty>
  items?: JsonSchemaProperty
  description?: string
}

const props = defineProps<{
  schema: Record<string, unknown>
  basePath: string
  clickable?: boolean
}>()

const emit = defineEmits<{
  fieldClick: [path: string]
}>()

const typeColors: Record<string, string> = {
  string: 'text-green-600 dark:text-green-400',
  number: 'text-blue-600 dark:text-blue-400',
  integer: 'text-blue-600 dark:text-blue-400',
  boolean: 'text-orange-600 dark:text-orange-400',
  object: 'text-gray-500 dark:text-gray-400',
  array: 'text-purple-600 dark:text-purple-400',
}

const typeDots: Record<string, string> = {
  string: 'bg-green-500',
  number: 'bg-blue-500',
  integer: 'bg-blue-500',
  boolean: 'bg-orange-500',
  object: 'bg-gray-400',
  array: 'bg-purple-500',
}

const parsed = computed(() => {
  const s = props.schema as JsonSchemaProperty
  if (!s.properties) return []
  return Object.entries(s.properties).map(([key, prop]) => ({
    key,
    type: prop.type ?? 'any',
    hasChildren: prop.type === 'object' && !!prop.properties,
    childSchema: prop as Record<string, unknown>,
    description: prop.description,
  }))
})

function onFieldClick(path: string) {
  emit('fieldClick', path)
}

function onDragStart(event: DragEvent, path: string) {
  event.dataTransfer?.setData('application/cactus-expression', path)
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'copy'
  }
}
</script>

<template>
  <div class="space-y-0.5 text-sm">
    <div
      v-for="field in parsed"
      :key="field.key"
    >
      <Collapsible v-if="field.hasChildren" :default-open="true">
        <CollapsibleTrigger class="flex w-full items-center gap-1.5 rounded px-1 py-0.5 hover:bg-accent">
          <ChevronRight class="h-3 w-3 shrink-0 transition-transform [[data-state=open]>&]:rotate-90" />
          <span
            class="inline-block h-2 w-2 rounded-full shrink-0"
            :class="typeDots[field.type] ?? 'bg-gray-400'"
          />
          <span class="font-medium">{{ field.key }}</span>
          <span :class="typeColors[field.type] ?? ''" class="text-xs ml-auto">{{ field.type }}</span>
        </CollapsibleTrigger>
        <CollapsibleContent class="pl-4">
          <SchemaTree
            :schema="field.childSchema"
            :base-path="`${basePath}.${field.key}`"
            :clickable="clickable"
            @field-click="onFieldClick"
          />
        </CollapsibleContent>
      </Collapsible>

      <div
        v-else
        class="flex items-center gap-1.5 rounded px-1 py-0.5"
        :class="clickable ? 'cursor-pointer hover:bg-accent' : ''"
        :draggable="clickable"
        @dragstart="onDragStart($event, `${basePath}.${field.key}`)"
        @click="clickable && onFieldClick(`${basePath}.${field.key}`)"
      >
        <span class="h-3 w-3 shrink-0" />
        <span
          class="inline-block h-2 w-2 rounded-full shrink-0"
          :class="typeDots[field.type] ?? 'bg-gray-400'"
        />
        <span class="font-medium">{{ field.key }}</span>
        <span :class="typeColors[field.type] ?? ''" class="text-xs ml-auto">{{ field.type }}</span>
      </div>
    </div>

    <div
      v-if="parsed.length === 0"
      class="px-1 py-2 text-xs text-muted-foreground"
    >
      No properties defined.
    </div>
  </div>
</template>

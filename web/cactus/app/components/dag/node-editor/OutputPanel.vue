<script setup lang="ts">
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '~/components/ui/tabs'
import { ScrollArea } from '~/components/ui/scroll-area'
import SchemaTree from './SchemaTree.vue'

const props = defineProps<{
  outputSchema?: Record<string, unknown>
}>()

const { t } = useI18n()

const jsonOutput = computed(() => {
  if (!props.outputSchema) return '{}'
  return JSON.stringify(props.outputSchema, null, 2)
})
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.output') || 'Output' }}</h3>
    </div>

    <Tabs default-value="schema" class="flex flex-1 flex-col overflow-hidden">
      <TabsList class="mx-4 mt-2 w-auto">
        <TabsTrigger value="schema">
          {{ t('nodeEditor.schema') || 'Schema' }}
        </TabsTrigger>
        <TabsTrigger value="json">
          JSON
        </TabsTrigger>
      </TabsList>

      <TabsContent value="schema" class="flex-1 overflow-hidden">
        <ScrollArea class="h-full">
          <div class="p-4">
            <SchemaTree
              v-if="outputSchema"
              :schema="outputSchema"
              base-path="$.output"
              :clickable="false"
            />
            <p
              v-else
              class="py-8 text-center text-sm text-muted-foreground"
            >
              No output schema defined.
            </p>
          </div>
        </ScrollArea>
      </TabsContent>

      <TabsContent value="json" class="flex-1 overflow-hidden">
        <ScrollArea class="h-full">
          <div class="p-4">
            <pre class="rounded-md bg-muted p-3 text-xs font-mono whitespace-pre-wrap">{{ jsonOutput }}</pre>
          </div>
        </ScrollArea>
      </TabsContent>
    </Tabs>
  </div>
</template>

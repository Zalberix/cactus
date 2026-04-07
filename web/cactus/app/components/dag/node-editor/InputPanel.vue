<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import type { Node, Edge } from '@vue-flow/core'
import type { StepData } from '~/composables/useDagEditor'
import { Input } from '~/components/ui/input'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'
import SchemaTree from './SchemaTree.vue'

const props = defineProps<{
  stepId: string
  allNodes: Node[]
  allEdges: Edge[]
}>()

const emit = defineEmits<{
  insertExpression: [expression: string]
}>()

const { t } = useI18n()
const searchQuery = ref('')

interface UpstreamNode {
  id: string
  label: string
  icon?: string
  color?: string
  outputSchema: Record<string, unknown>
}

const upstreamNodes = computed<UpstreamNode[]>(() => {
  const upstreamEdges = props.allEdges.filter(e => e.target === props.stepId)
  const nodes: UpstreamNode[] = []

  for (const edge of upstreamEdges) {
    const node = props.allNodes.find(n => n.id === edge.source)
    if (!node) continue
    const data = node.data as StepData
    if (!data.outputSchema) continue

    nodes.push({
      id: edge.source,
      label: data.label,
      icon: data.workTypeMeta?.icon,
      color: data.workTypeMeta?.color ?? '#607d8b',
      outputSchema: data.outputSchema,
    })
  }

  return nodes
})

function onFieldClick(path: string) {
  emit('insertExpression', path)
}
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.input') || 'Input' }}</h3>
      <p class="text-xs text-muted-foreground">
        {{ t('nodeEditor.upstream') || 'Upstream Outputs' }}
      </p>
    </div>

    <!-- Search -->
    <div class="px-4 py-2 shrink-0">
      <div class="relative">
        <Search class="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          placeholder="Search fields..."
          class="h-8 pl-8 text-xs"
        />
      </div>
    </div>

    <ScrollArea class="flex-1">
      <div class="px-4 pb-4 space-y-3">
        <div v-if="upstreamNodes.length === 0" class="py-8 text-center">
          <p class="text-sm text-muted-foreground">
            {{ t('nodeEditor.noUpstream') || 'No upstream dependencies.' }}
          </p>
        </div>

        <div v-for="upstream in upstreamNodes" :key="upstream.id">
          <div class="mb-2 flex items-center gap-2">
            <div
              class="h-2 w-2 rounded-full shrink-0"
              :style="{ backgroundColor: upstream.color }"
            />
            <span class="text-xs font-medium">
              {{ upstream.label }}
            </span>
            <span class="text-[10px] text-muted-foreground">
              #{{ upstream.id }}
            </span>
          </div>

          <SchemaTree
            :schema="upstream.outputSchema"
            :base-path="`$.steps.${upstream.id}.output`"
            :clickable="true"
            @field-click="onFieldClick"
          />

          <Separator class="mt-3" />
        </div>

        <!-- Message context -->
        <div>
          <div class="mb-2 flex items-center gap-2">
            <div class="h-2 w-2 rounded-full shrink-0 bg-amber-500" />
            <span class="text-xs font-medium">Message</span>
          </div>
          <div
            class="flex items-center gap-1.5 rounded px-1 py-0.5 text-sm cursor-pointer hover:bg-accent"
            @click="onFieldClick('$.message.value')"
          >
            <span class="h-3 w-3 shrink-0" />
            <span class="inline-block h-2 w-2 rounded-full shrink-0 bg-gray-400" />
            <span class="font-medium">value</span>
            <span class="text-xs text-muted-foreground ml-auto">object</span>
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

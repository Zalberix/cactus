<script setup lang="ts">
import { getSmoothStepPath, BaseEdge, EdgeLabelRenderer } from '@vue-flow/core'
import type { EdgeProps } from '@vue-flow/core'
import { X } from 'lucide-vue-next'

const props = defineProps<EdgeProps>()
const emit = defineEmits(['remove'])

const isHovered = ref(false)

const edgeColor = computed(() => {
  if (props.selected) return '#64748b'
  return '#94a3b8'
})

const pathParams = computed(() => {
  return getSmoothStepPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition,
    borderRadius: 8,
  })
})

const edgePath = computed(() => pathParams.value[0])
const labelX = computed(() => pathParams.value[1])
const labelY = computed(() => pathParams.value[2])
</script>

<template>
  <BaseEdge
    :id="id"
    :path="edgePath"
    :interaction-width="24"
    :style="{ stroke: edgeColor, strokeWidth: props.selected ? 3 : 2 }"
    @mouseenter="isHovered = true"
    @mouseleave="isHovered = false"
  />
  <EdgeLabelRenderer>
    <div
      class="vue-flow__edge-label nodrag nopan pointer-events-auto"
      :style="{
        position: 'absolute',
        transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
      }"
      @mouseenter="isHovered = true"
      @mouseleave="isHovered = false"
    >
      <button
        v-if="isHovered || selected"
        type="button"
        aria-label="Remove connection"
        title="Remove connection"
        class="flex h-5 w-5 items-center justify-center rounded-full border border-slate-300 bg-background text-muted-foreground shadow-sm hover:bg-destructive hover:text-destructive-foreground hover:border-destructive transition-colors"
        @click.stop="emit('remove')"
      >
        <X class="h-3 w-3" />
      </button>
    </div>
  </EdgeLabelRenderer>
</template>

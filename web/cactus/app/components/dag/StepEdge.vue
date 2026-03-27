<script setup lang="ts">
import { getBezierPath, BaseEdge, EdgeLabelRenderer } from '@vue-flow/core'
import type { EdgeProps } from '@vue-flow/core'
import { X } from 'lucide-vue-next'

const props = defineProps<EdgeProps>()
const emit = defineEmits(['remove'])

const edgeColors: Record<string, string> = {
  success: '#22c55e',
  failure: '#ef4444',
  skip: '#9ca3af',
}

const labelColors: Record<string, string> = {
  success: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  failure: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
  skip: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300',
}

const isHovered = ref(false)

const outcome = computed(() => props.sourceHandleId ?? 'success')

const pathParams = computed(() => {
  return getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition,
  })
})

const edgePath = computed(() => pathParams.value[0])
const labelX = computed(() => pathParams.value[1])
const labelY = computed(() => pathParams.value[2])
const strokeColor = computed(() => edgeColors[outcome.value] ?? edgeColors.success)
const labelClass = computed(() => labelColors[outcome.value] ?? labelColors.success)
</script>

<template>
  <BaseEdge
    :id="id"
    :path="edgePath"
    :style="{ stroke: strokeColor, strokeWidth: 2 }"
    @mouseenter="isHovered = true"
    @mouseleave="isHovered = false"
  />
  <EdgeLabelRenderer>
    <div
      class="vue-flow__edge-label nodrag nopan pointer-events-auto flex items-center gap-1"
      :style="{
        position: 'absolute',
        transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
      }"
      @mouseenter="isHovered = true"
      @mouseleave="isHovered = false"
    >
      <span
        class="rounded px-1.5 py-0.5 text-[10px] font-medium"
        :class="labelClass"
      >
        {{ outcome }}
      </span>
      <button
        v-if="isHovered"
        class="flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-opacity"
        @click.stop="emit('remove')"
      >
        <X class="h-3 w-3" />
      </button>
    </div>
  </EdgeLabelRenderer>
</template>

<script setup lang="ts">
import { VueFlow, useVueFlow } from '@vue-flow/core'
import type { Node, Edge, Connection, NodeDragEvent, NodeMouseEvent, EdgeMouseEvent } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import StepNode from './StepNode.vue'
import StepEdge from './StepEdge.vue'

const GRID_SIZE = 20

const props = withDefaults(defineProps<{
  mode?: 'edit' | 'view'
  nodes: Node[]
  edges: Edge[]
}>(), {
  mode: 'edit',
})

const emit = defineEmits<{
  connect: [params: Connection]
  nodeDragStop: [nodeId: string, position: { x: number; y: number }]
  nodeClick: [nodeId: string]
  nodeDoubleClick: [nodeId: string]
  edgeClick: [edgeId: string]
  removeEdge: [edgeId: string]
  drop: [stepType: string, workTypeId: number | undefined, workTypeCode: string | undefined, position: { x: number; y: number }, name: string | undefined]
  deleteSelected: []
}>()

const nodeTypes = {
  step: StepNode,
}

const edgeTypes = {
  step: StepEdge,
}

const { project, fitView } = useVueFlow()

const isEdit = computed(() => props.mode === 'edit')

function snapToGrid(val: number): number {
  return Math.round(val / GRID_SIZE) * GRID_SIZE
}

function onConnect(params: Connection) {
  if (!isEdit.value) return
  emit('connect', params)
}

function onNodeDragStop(event: NodeDragEvent) {
  if (!isEdit.value) return
  const pos = {
    x: snapToGrid(event.node.position.x),
    y: snapToGrid(event.node.position.y),
  }
  emit('nodeDragStop', event.node.id, pos)
}

function onNodeClick(event: NodeMouseEvent) {
  emit('nodeClick', event.node.id)
}

function onNodeDoubleClick(event: NodeMouseEvent) {
  emit('nodeDoubleClick', event.node.id)
}

function onEdgeClick(event: EdgeMouseEvent) {
  emit('edgeClick', event.edge.id)
}

function onEdgeRemove(edgeId: string) {
  emit('removeEdge', edgeId)
}

function onDragOver(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

function onDrop(event: DragEvent) {
  if (!isEdit.value || !event.dataTransfer) return

  const raw = event.dataTransfer.getData('application/cactus-step')
  if (!raw) return

  const data = JSON.parse(raw) as {
    stepType: string
    workTypeId?: number
    workTypeCode?: string
    name?: string
  }

  const projected = project({
    x: event.clientX,
    y: event.clientY,
  })

  const position = {
    x: snapToGrid(projected.x),
    y: snapToGrid(projected.y),
  }

  emit('drop', data.stepType, data.workTypeId, data.workTypeCode, position, data.name)
}

function onKeyDown(event: KeyboardEvent) {
  if (!isEdit.value) return
  if (event.key === 'Delete' || event.key === 'Backspace') {
    emit('deleteSelected')
  }
}

onMounted(() => {
  nextTick(() => {
    fitView({ padding: 0.2 })
  })
})
</script>

<template>
  <div
    class="relative h-full w-full"
    @keydown="onKeyDown"
    tabindex="0"
  >
    <VueFlow
      :nodes="nodes"
      :edges="edges"
      :node-types="nodeTypes"
      :edge-types="edgeTypes"
      :nodes-draggable="isEdit"
      :nodes-connectable="isEdit"
      :elements-selectable="true"
      :zoom-on-scroll="true"
      :pan-on-drag="true"
      :auto-connect="false"
      :snap-to-grid="true"
      :snap-grid="[GRID_SIZE, GRID_SIZE]"
      fit-view-on-init
      @connect="onConnect"
      @node-drag-stop="onNodeDragStop"
      @node-click="onNodeClick"
      @node-double-click="onNodeDoubleClick"
      @edge-click="onEdgeClick"
      @dragover="onDragOver"
      @drop="onDrop"
    >
      <Background :gap="GRID_SIZE" />
      <Controls />

      <template #node-step="nodeProps">
        <StepNode v-bind="nodeProps" />
      </template>

      <template #edge-step="edgeProps">
        <StepEdge
          v-bind="edgeProps"
          @remove="onEdgeRemove(edgeProps.id)"
        />
      </template>
    </VueFlow>

    <!-- Empty state -->
    <div
      v-if="nodes.length === 0"
      class="absolute inset-0 flex items-center justify-center pointer-events-none"
    >
      <div class="text-center text-muted-foreground">
        <p class="text-lg font-medium">{{ $t('empty.dagCanvas.heading') }}</p>
        <p class="mt-1 text-sm">{{ $t('empty.dagCanvas.body') }}</p>
      </div>
    </div>
  </div>
</template>

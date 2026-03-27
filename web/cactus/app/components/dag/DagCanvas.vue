<script setup lang="ts">
import { VueFlow, useVueFlow } from '@vue-flow/core'
import type { Node, Edge, Connection, NodeDragEvent, NodeMouseEvent, EdgeMouseEvent } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import StepNode from './StepNode.vue'
import StepEdge from './StepEdge.vue'

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
  edgeClick: [edgeId: string]
  removeEdge: [edgeId: string]
  drop: [stepType: string, workTypeId: number | undefined, position: { x: number; y: number }, name: string | undefined]
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

function onConnect(params: Connection) {
  if (!isEdit.value) return
  emit('connect', params)
}

function onNodeDragStop(event: NodeDragEvent) {
  if (!isEdit.value) return
  emit('nodeDragStop', event.node.id, event.node.position)
}

function onNodeClick(event: NodeMouseEvent) {
  emit('nodeClick', event.node.id)
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
    name?: string
  }

  const position = project({
    x: event.clientX,
    y: event.clientY,
  })

  emit('drop', data.stepType, data.workTypeId, position, data.name)
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
      fit-view-on-init
      @connect="onConnect"
      @node-drag-stop="onNodeDragStop"
      @node-click="onNodeClick"
      @edge-click="onEdgeClick"
      @dragover="onDragOver"
      @drop="onDrop"
    >
      <Background />
      <Controls />

      <!-- Empty state overlay -->
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

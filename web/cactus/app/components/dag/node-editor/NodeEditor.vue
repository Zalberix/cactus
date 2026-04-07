<script setup lang="ts">
import type { Node, Edge } from '@vue-flow/core'
import type { StepData } from '~/composables/useDagEditor'
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap, X,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import {
  Dialog,
  DialogContent,
} from '~/components/ui/dialog'
import InputPanel from './InputPanel.vue'
import ParametersPanel from './ParametersPanel.vue'
import OutputPanel from './OutputPanel.vue'

const iconMap: Record<string, Component> = {
  'mail': Mail,
  'message-square': MessageSquare,
  'bell': Bell,
  'workflow': Workflow,
  'git-branch': GitBranch,
  'clock': Clock,
  'split': Split,
  'zap': Zap,
}

const props = defineProps<{
  open: boolean
  nodeId: string | null
  allNodes: Node[]
  allEdges: Edge[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [nodeId: string, config: Record<string, unknown>, inputMapping: Record<string, string>]
}>()

const { t } = useI18n()

const dialogOpen = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val),
})

const currentNode = computed(() => {
  if (!props.nodeId) return null
  return props.allNodes.find(n => n.id === props.nodeId) ?? null
})

const stepData = computed<StepData | null>(() => {
  if (!currentNode.value) return null
  return currentNode.value.data as StepData
})

const nodeIcon = computed(() => {
  if (!stepData.value) return Workflow
  const metaIcon = stepData.value.workTypeMeta?.icon
  if (metaIcon && iconMap[metaIcon]) return iconMap[metaIcon]
  if (stepData.value.stepType === 'control') return GitBranch
  return Workflow
})

const accentColor = computed(() => stepData.value?.workTypeMeta?.color ?? '#607d8b')

function onSave(config: Record<string, unknown>, inputMapping: Record<string, string>) {
  if (!props.nodeId) return
  emit('save', props.nodeId, config, inputMapping)
}

const parametersRef = ref<InstanceType<typeof ParametersPanel> | null>(null)

function onInsertExpression(expression: string) {
  parametersRef.value?.insertExpression(expression)
}
</script>

<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent class="max-w-[95vw] h-[90vh] flex flex-col p-0 gap-0">
      <!-- Header -->
      <div class="flex items-center gap-3 border-b px-5 py-3 shrink-0">
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
          :style="{ backgroundColor: accentColor + '20', color: accentColor }"
        >
          <component :is="nodeIcon" class="h-5 w-5" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-sm font-semibold truncate">
            {{ stepData?.label ?? 'Step' }}
          </div>
          <div class="flex items-center gap-1.5">
            <Badge variant="secondary" class="h-5 px-1.5 text-[10px]">
              {{ stepData?.stepType === 'control' ? 'Control' : 'Task' }}
            </Badge>
            <Badge
              v-if="stepData?.workTypeCode"
              variant="outline"
              class="h-5 px-1.5 text-[10px]"
            >
              {{ stepData.workTypeCode }}
            </Badge>
          </div>
        </div>
        <Button variant="ghost" size="icon" class="h-8 w-8 shrink-0" @click="dialogOpen = false">
          <X class="h-4 w-4" />
        </Button>
      </div>

      <!-- Three panels -->
      <div v-if="stepData && nodeId" class="flex flex-1 overflow-hidden">
        <div class="w-1/4 border-r overflow-hidden">
          <InputPanel
            :step-id="nodeId"
            :all-nodes="allNodes"
            :all-edges="allEdges"
            @insert-expression="onInsertExpression"
          />
        </div>
        <div class="w-1/2 overflow-hidden">
          <ParametersPanel
            ref="parametersRef"
            :step-data="stepData"
            @save="onSave"
          />
        </div>
        <div class="w-1/4 border-l overflow-hidden">
          <OutputPanel :output-schema="stepData.outputSchema" />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

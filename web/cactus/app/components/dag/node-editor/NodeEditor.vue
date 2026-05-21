<script setup lang="ts">
import type { Node, Edge } from '@vue-flow/core'
import type { StepData } from '~/composables/useDagEditor'
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import { Badge } from '~/components/ui/badge'
import {
  Dialog,
  DialogContent,
} from '~/components/ui/dialog'
import InputPanel from './InputPanel.vue'
import ParametersPanel from './ParametersPanel.vue'
import ControlParametersPanel from './ControlParametersPanel.vue'
import { useControlSteps } from '~/composables/useControlSteps'

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

const controlSteps = useControlSteps()

const props = defineProps<{
  open: boolean
  nodeId: string | null
  workflowId: number
  versionId: number | null
  allNodes: Node[]
  allEdges: Edge[]
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saveSettings: [nodeId: string, settingsData: Record<string, unknown>]
  saveControlSettings: [nodeId: string, settingsData: Record<string, unknown>]
  saveInputMapping: [nodeId: string, inputMapping: Array<{ target: string, source: string }>]
  workflowInputsChanged: []
}>()

const dialogOpen = computed({
  get: () => props.open,
  set: val => emit('update:open', val),
})

const currentNode = computed(() => {
  if (!props.nodeId) return null
  return props.allNodes.find(n => n.id === props.nodeId) ?? null
})

const stepData = computed<StepData | null>(() => {
  if (!currentNode.value) return null
  return currentNode.value.data as StepData
})

const isControl = computed(() =>
  stepData.value?.stepType === 'control' && stepData.value.controlKind !== 'start',
)

const controlDefinition = computed(() =>
  isControl.value
    ? controlSteps.find(item => item.kind === stepData.value?.controlKind)
    : undefined,
)

const nodeIcon = computed(() => {
  if (!stepData.value) return Workflow
  const controlIcon = controlDefinition.value?.icon
  if (controlIcon && iconMap[controlIcon]) return iconMap[controlIcon]
  const metaIcon = stepData.value.workTypeMeta?.icon
  if (metaIcon && iconMap[metaIcon]) return iconMap[metaIcon]
  return Workflow
})

const accentColor = computed(() =>
  controlDefinition.value?.color ?? stepData.value?.workTypeMeta?.color ?? '#607d8b',
)

const stepTypeName = computed(() =>
  controlDefinition.value?.name ?? stepData.value?.workTypeName ?? 'Task',
)

const stepTypeCode = computed(() =>
  stepData.value?.workTypeCode ?? stepData.value?.controlKind ?? '',
)

function onSaveSettings(settingsData: Record<string, unknown>) {
  if (!props.nodeId) return
  emit('saveSettings', props.nodeId, settingsData)
}

function onSaveControlSettings(settingsData: Record<string, unknown>) {
  if (!props.nodeId) return
  emit('saveControlSettings', props.nodeId, settingsData)
}

function onSaveInputMapping(inputMapping: Array<{ target: string, source: string }>) {
  if (!props.nodeId) return
  emit('saveInputMapping', props.nodeId, inputMapping)
}

const parametersRef = ref<InstanceType<typeof ParametersPanel> | null>(null)
const controlParametersRef = ref<InstanceType<typeof ControlParametersPanel> | null>(null)
const inputPanelRef = ref<InstanceType<typeof InputPanel> | null>(null)

function onInsertExpression(expression: string) {
  if (isControl.value) {
    controlParametersRef.value?.insertExpression(expression)
    return
  }
  parametersRef.value?.insertExpression(expression)
}

function onCreateWorkflowInput(field: string, property: Record<string, unknown>) {
  inputPanelRef.value?.createWorkflowInputFromField(field, property, (expression: string) => {
    if (isControl.value) {
      controlParametersRef.value?.insertExpression(expression)
      return
    }
    parametersRef.value?.insertExpression(expression)
  })
}
</script>

<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent class="max-w-[95vw] h-[90vh] flex flex-col p-0 gap-0">
      <div class="flex items-center gap-3 border-b px-5 py-3 shrink-0">
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
          :style="{ backgroundColor: accentColor + '20', color: accentColor }"
        >
          <component :is="nodeIcon" class="h-5 w-5" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="text-sm font-semibold truncate">
            {{ stepData?.label ?? 'Task' }}
          </div>
          <div class="flex items-center gap-1.5">
            <Badge variant="secondary" class="h-5 px-1.5 text-[10px]">
              {{ stepTypeName }}
            </Badge>
            <Badge
              v-if="stepTypeCode"
              variant="outline"
              class="h-5 px-1.5 text-[10px]"
            >
              {{ stepTypeCode }}
            </Badge>
          </div>
        </div>
      </div>

      <div v-if="stepData && nodeId" class="flex flex-1 overflow-hidden">
        <div class="w-1/3 border-r overflow-hidden">
          <InputPanel
            ref="inputPanelRef"
            :step-id="nodeId"
            :workflow-id="workflowId"
            :version-id="versionId"
            :all-nodes="allNodes"
            :all-edges="allEdges"
            @insert-expression="onInsertExpression"
            @workflow-inputs-changed="emit('workflowInputsChanged')"
          />
        </div>
        <div class="w-2/3 overflow-hidden">
          <ParametersPanel
            v-if="!isControl"
            ref="parametersRef"
            :step-data="stepData"
            @create-workflow-input="onCreateWorkflowInput"
            @save-settings="onSaveSettings"
            @save-input-mapping="onSaveInputMapping"
          />
          <ControlParametersPanel
            v-else
            ref="controlParametersRef"
            :step-data="stepData"
            @save-control-settings="onSaveControlSettings"
          />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

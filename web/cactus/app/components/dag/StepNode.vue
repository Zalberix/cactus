<script setup lang="ts">
import { Handle, Position, useNode } from '@vue-flow/core'
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap, Radio, AlertCircle, Trash2, MoreHorizontal,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import { controlOutcomesForStep } from '~/composables/control-outcomes'
import { useControlSteps } from '~/composables/useControlSteps'

const { node } = useNode()
const { t } = useI18n()
const controlSteps = useControlSteps()
const emit = defineEmits<{
  delete: [nodeId: string]
}>()

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

const isControl = computed(() => node.data.stepType === 'control')
const isStart = computed(() => node.data.controlKind === 'start')

const icon = computed(() => {
  const metaIcon = node.data.workTypeMeta?.icon
  if (metaIcon && iconMap[metaIcon]) return iconMap[metaIcon]
  if (isStart.value) return Radio
  if (node.data.stepType === 'control') return GitBranch
  return Workflow
})

const accentColor = computed(() => {
  if (isStart.value) return '#0f766e'
  const control = controlSteps.find(item => item.kind === node.data.controlKind)
  return control?.color ?? node.data.workTypeMeta?.color ?? '#607d8b'
})

const outputHandles = computed(() => {
  if (isStart.value) {
    return [{ id: 'success', label: '', color: '#22c55e' }]
  }
  if (isControl.value && node.data.controlKind) {
    return controlOutcomesForStep(node.data)
  }
  return [{ id: 'success', label: '', color: '#22c55e' }]
})

const label = computed(() => node.data.label ?? 'Step')
const nodeErrors = computed(() => node.data.validationErrors ?? [])
const hasValidationErrors = computed(() => nodeErrors.value.length > 0)
const canShowActions = computed(() => !isStart.value)

function onDelete(event: MouseEvent) {
  event.stopPropagation()
  if (isStart.value) return
  emit('delete', node.id)
}
</script>

<template>
  <div class="group/node relative flex w-[168px] flex-col items-center select-none">
    <div
      v-if="canShowActions"
      class="pointer-events-auto absolute left-1/2 top-0 z-20 flex h-10 -translate-x-1/2 -translate-y-full items-end justify-center gap-1 pb-1 opacity-0 transition-opacity group-hover/node:opacity-100 hover:opacity-100 focus-within:opacity-100"
      data-testid="node-hover-actions"
    >
      <button
        type="button"
        data-testid="delete-node"
        class="flex h-7 w-7 items-center justify-center rounded-md border border-slate-300 bg-background text-muted-foreground shadow-sm hover:border-destructive hover:bg-destructive hover:text-destructive-foreground"
        :title="t('editor.deleteStep')"
        :aria-label="t('editor.deleteStep')"
        @click="onDelete"
      >
        <Trash2 class="h-3.5 w-3.5" />
      </button>

      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <button
            type="button"
            data-testid="more-node"
            class="flex h-7 w-7 items-center justify-center rounded-md border border-slate-300 bg-background text-muted-foreground shadow-sm hover:bg-accent hover:text-accent-foreground"
            :title="t('common.actions')"
            :aria-label="t('common.actions')"
            @click.stop
          >
            <MoreHorizontal class="h-4 w-4" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="center" class="w-36">
          <DropdownMenuItem @select.prevent>
            {{ t('editor.renameStep') }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <div
      data-testid="step-node-block"
      class="relative flex h-[104px] w-[144px] items-center justify-center rounded-lg border-2 border-slate-300 bg-background shadow-sm transition-colors hover:shadow-md"
      :class="[
        node.selected ? 'ring-2 ring-slate-400/30' : '',
        hasValidationErrors ? 'border-red-500 ring-2 ring-red-500/20' : '',
      ]"
    >
      <Handle
        v-if="!isStart"
        type="target"
        :position="Position.Left"
        class="!h-4 !w-4 !border-2 !border-background !bg-gray-400 !-left-2"
      />

      <div
        class="flex h-12 w-12 items-center justify-center rounded-md"
        :style="{ backgroundColor: accentColor + '20', color: accentColor }"
      >
        <component :is="icon" class="h-6 w-6" />
      </div>

      <div
        v-if="hasValidationErrors"
        class="group/error absolute bottom-1 right-1 z-10"
      >
        <button
          type="button"
          data-testid="validation-indicator"
          class="flex h-6 w-6 items-center justify-center rounded-full border border-red-200 bg-red-50 text-red-600 shadow-sm"
          :aria-label="t('editor.validationErrors')"
          @click.stop
        >
          <AlertCircle class="h-3.5 w-3.5" />
        </button>
        <div
          data-testid="validation-tooltip"
          class="pointer-events-none absolute bottom-7 right-0 hidden w-64 rounded-md border bg-popover p-2 text-xs text-popover-foreground shadow-md group-hover/error:block"
        >
          <div v-for="(error, index) in nodeErrors" :key="index" class="py-0.5">
            {{ error.message }}
          </div>
        </div>
      </div>

      <Handle
        v-for="(handle, idx) in outputHandles"
        :key="handle.id"
        type="source"
        :id="handle.id"
        :position="Position.Right"
        class="!h-4 !w-4 !border-2 !border-background !-right-2"
        :style="{
          backgroundColor: handle.color,
          top: outputHandles.length === 1
            ? '50%'
            : `${24 + (idx * 52 / Math.max(outputHandles.length - 1, 1))}%`,
        }"
      />

      <template v-if="outputHandles.length > 1">
        <div
          v-for="(handle, idx) in outputHandles"
          :key="`label-${handle.id}`"
          data-testid="output-handle-label"
          class="absolute left-[calc(100%+14px)] max-w-[88px] truncate rounded-sm bg-background/90 px-1 py-0.5 text-[10px] font-medium leading-none shadow-sm ring-1 ring-slate-200 pointer-events-none"
          :style="{
            color: handle.color,
            top: `${24 + (idx * 52 / Math.max(outputHandles.length - 1, 1))}%`,
            transform: 'translateY(calc(-100% - 6px))',
          }"
        >
          {{ handle.label }}
        </div>
      </template>
    </div>

    <div
      data-testid="step-node-label"
      class="mt-2 max-w-[180px] truncate text-center text-sm font-semibold leading-tight text-foreground"
      :title="label"
    >
      {{ label }}
    </div>
  </div>
</template>

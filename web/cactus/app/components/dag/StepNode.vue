<script setup lang="ts">
import { Handle, Position, useNode } from '@vue-flow/core'
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap, Radio, AlertCircle,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import { useControlSteps } from '~/composables/useControlSteps'

const { node } = useNode()
const controlSteps = useControlSteps()

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
    const definition = controlSteps.find(control => control.kind === node.data.controlKind)
    return definition?.handles ?? [{ id: 'success', label: '', color: '#22c55e' }]
  }
  return [{ id: 'success', label: '', color: '#22c55e' }]
})

const statusIndicator = computed(() => {
  const s = node.data.status ?? 'pending'
  const map: Record<string, string> = {
    pending: 'bg-gray-300 dark:bg-gray-600',
    running: 'bg-blue-500 animate-pulse',
    done: 'bg-green-500',
    error: 'bg-red-500',
  }
  return map[s] ?? map.pending
})

const label = computed(() => node.data.label ?? 'Step')
const subtitle = computed(() => {
  if (isStart.value) return 'system_message'
  return node.data.workTypeCode ?? node.data.controlKind ?? ''
})
const nodeErrors = computed(() => node.data.validationErrors ?? [])
const hasValidationErrors = computed(() => nodeErrors.value.length > 0)
</script>

<template>
  <div
    class="relative flex items-stretch rounded-lg border bg-background shadow-sm transition-shadow hover:shadow-md cursor-pointer select-none"
    :class="[
      node.selected ? 'ring-2 ring-primary shadow-md' : '',
      hasValidationErrors ? 'ring-2 ring-orange-500' : '',
    ]"
    style="min-width: 200px;"
  >
    <div v-if="hasValidationErrors" class="group absolute right-1 top-1 z-10">
      <button
        type="button"
        class="flex h-6 w-6 items-center justify-center rounded bg-background text-orange-600 opacity-0 shadow-sm ring-1 ring-border transition-opacity group-hover:opacity-100"
      >
        <AlertCircle class="h-3.5 w-3.5" />
      </button>
      <div class="pointer-events-none absolute right-0 top-7 hidden w-64 rounded-md border bg-popover p-2 text-xs text-popover-foreground shadow-md group-hover:block">
        <div v-for="(error, index) in nodeErrors" :key="index" class="py-0.5">
          {{ error.message }}
        </div>
      </div>
    </div>
    <div
      class="w-1 shrink-0 rounded-l-lg"
      :style="{ backgroundColor: accentColor }"
    />

    <Handle
      v-if="!isStart"
      type="target"
      :position="Position.Left"
      class="!w-3 !h-3 !border-2 !border-background !bg-gray-400 !-left-1.5"
    />

    <div class="flex items-center gap-3 px-3 py-2.5 min-w-0 flex-1">
      <div
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
        :style="{ backgroundColor: accentColor + '20', color: accentColor }"
      >
        <component :is="icon" class="h-4 w-4" />
      </div>

      <div class="min-w-0 flex-1">
        <div class="truncate text-sm font-medium leading-tight">{{ label }}</div>
        <div
          v-if="subtitle"
          class="truncate text-[11px] text-muted-foreground leading-tight"
        >
          {{ subtitle }}
        </div>
      </div>

      <div
        class="h-2 w-2 shrink-0 rounded-full"
        :class="statusIndicator"
      />
    </div>

    <div class="relative shrink-0 flex flex-col justify-center" style="width: 6px;">
      <Handle
        v-for="(handle, idx) in outputHandles"
        :key="handle.id"
        type="source"
        :id="handle.id"
        :position="Position.Right"
        class="!w-3 !h-3 !border-2 !border-background !-right-1.5"
        :style="{
          backgroundColor: handle.color,
          top: outputHandles.length === 1
            ? '50%'
            : `${20 + (idx * 60 / Math.max(outputHandles.length - 1, 1))}%`,
        }"
      />

      <template v-if="outputHandles.length > 1">
        <div
          v-for="(handle, idx) in outputHandles"
          :key="`label-${handle.id}`"
          class="absolute text-[9px] font-medium leading-none pointer-events-none"
          :style="{
            color: handle.color,
            right: '10px',
            top: `${20 + (idx * 60 / Math.max(outputHandles.length - 1, 1))}%`,
            transform: 'translateY(-50%)',
          }"
        >
          {{ handle.label }}
        </div>
      </template>
    </div>
  </div>
</template>

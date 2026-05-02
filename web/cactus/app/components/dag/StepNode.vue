<script setup lang="ts">
import { Handle, Position, useNode } from '@vue-flow/core'
import {
  Mail, MessageSquare, Bell, Workflow, GitBranch,
  Clock, Split, Zap, Radio,
} from 'lucide-vue-next'
import type { Component } from 'vue'

const { node } = useNode()

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

// Control node output handles
const controlHandles: Record<string, Array<{ id: string; label: string; color: string }>> = {
  condition: [
    { id: 'true', label: 'true', color: '#22c55e' },
    { id: 'false', label: 'false', color: '#ef4444' },
  ],
  switch: [
    { id: 'case0', label: '0', color: '#3b82f6' },
    { id: 'case1', label: '1', color: '#8b5cf6' },
    { id: 'default', label: '∗', color: '#6b7280' },
  ],
  delay: [
    { id: 'continue', label: '', color: '#22c55e' },
  ],
}

const icon = computed(() => {
  const metaIcon = node.data.workTypeMeta?.icon
  if (metaIcon && iconMap[metaIcon]) return iconMap[metaIcon]
  if (isStart.value) return Radio
  if (node.data.stepType === 'control') return GitBranch
  return Workflow
})

const accentColor = computed(() => {
  if (isStart.value) return '#0f766e'
  return node.data.workTypeMeta?.color ?? '#607d8b'
})

const isControl = computed(() => node.data.stepType === 'control')
const isStart = computed(() => node.data.controlKind === 'start')

const outputHandles = computed(() => {
  if (isStart.value) {
    return [{ id: 'success', label: '', color: '#22c55e' }]
  }
  if (isControl.value && node.data.controlKind) {
    return controlHandles[node.data.controlKind] ?? [{ id: 'success', label: '', color: '#22c55e' }]
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
</script>

<template>
  <div
    class="relative flex items-stretch rounded-lg border bg-background shadow-sm transition-shadow hover:shadow-md cursor-pointer select-none"
    :class="node.selected ? 'ring-2 ring-primary shadow-md' : ''"
    style="min-width: 200px;"
  >
    <!-- Color accent bar (left) -->
    <div
      class="w-1 shrink-0 rounded-l-lg"
      :style="{ backgroundColor: accentColor }"
    />

    <!-- Input handle (left side) -->
    <Handle
      v-if="!isStart"
      type="target"
      :position="Position.Left"
      class="!w-3 !h-3 !border-2 !border-background !bg-gray-400 !-left-1.5"
    />

    <!-- Content -->
    <div class="flex items-center gap-3 px-3 py-2.5 min-w-0 flex-1">
      <!-- Icon circle -->
      <div
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
        :style="{ backgroundColor: accentColor + '20', color: accentColor }"
      >
        <component :is="icon" class="h-4 w-4" />
      </div>

      <!-- Text -->
      <div class="min-w-0 flex-1">
        <div class="truncate text-sm font-medium leading-tight">{{ label }}</div>
        <div
          v-if="subtitle"
          class="truncate text-[11px] text-muted-foreground leading-tight"
        >
          {{ subtitle }}
        </div>
      </div>

      <!-- Status dot -->
      <div
        class="h-2 w-2 shrink-0 rounded-full"
        :class="statusIndicator"
      />
    </div>

    <!-- Output handles (right side) -->
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

      <!-- Handle labels for multi-output controls -->
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

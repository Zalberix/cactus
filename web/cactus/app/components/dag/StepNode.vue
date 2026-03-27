<script setup lang="ts">
import { Handle, Position, useNode } from '@vue-flow/core'
import { Mail, MessageSquare, Bell, Workflow, GitBranch } from 'lucide-vue-next'
import type { Component } from 'vue'

const { node } = useNode()

const statusColors: Record<string, string> = {
  pending: 'border-muted-foreground bg-muted',
  running: 'border-blue-500 bg-blue-50 dark:bg-blue-950 animate-pulse',
  done: 'border-green-500 bg-green-50 dark:bg-green-950',
  error: 'border-red-500 bg-red-50 dark:bg-red-950',
}

const workTypeIcons: Record<string, Component> = {
  email: Mail,
  smtp: Mail,
  sms: MessageSquare,
  push: Bell,
  telegram: MessageSquare,
}

const icon = computed(() => {
  if (node.data.stepType === 'control') return GitBranch

  const wtName = (node.data.workTypeName ?? '').toLowerCase()
  return workTypeIcons[wtName] ?? Workflow
})

const nodeClass = computed(() => {
  const base = 'rounded-lg border-2 p-3 min-w-[180px] shadow-sm transition-colors duration-300 cursor-pointer'
  const statusClass = statusColors[node.data.status] ?? statusColors.pending
  return `${base} ${statusClass}`
})
</script>

<template>
  <div :class="nodeClass">
    <Handle
      type="target"
      :position="Position.Top"
    />

    <div class="flex items-center gap-2">
      <component
        :is="icon"
        class="size-5 shrink-0"
      />
      <span class="font-semibold text-sm truncate">{{ node.data.label }}</span>
    </div>

    <Handle
      type="source"
      id="success"
      :position="Position.Bottom"
      class="!bg-green-500"
    />
    <Handle
      type="source"
      id="failure"
      :position="Position.Right"
      class="!bg-red-500"
    />
    <Handle
      type="source"
      id="skip"
      :position="Position.Left"
      class="!bg-gray-400"
    />
  </div>
</template>

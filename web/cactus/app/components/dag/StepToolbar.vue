<script setup lang="ts">
import {
  Mail,
  MessageSquare,
  Bell,
  Workflow,
  GitBranch,
  PanelLeftClose,
  PanelLeftOpen,
  GripVertical,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import type { WorkType } from '~/composables/useVersions'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Button } from '~/components/ui/button'

const emit = defineEmits<{
  addStep: [stepType: string, workTypeId: number | undefined, position: { x: number; y: number }]
}>()

const { t } = useI18n()

const { fetchWorkTypes } = useVersions()

const workTypes = ref<WorkType[]>([])
const collapsed = ref(false)

const workTypeIcons: Record<string, Component> = {
  email: Mail,
  smtp: Mail,
  sms: MessageSquare,
  push: Bell,
  telegram: MessageSquare,
}

interface ToolbarItem {
  id: string
  name: string
  icon: Component
  stepType: string
  workTypeId?: number
}

const items = computed<ToolbarItem[]>(() => {
  const list: ToolbarItem[] = [
    {
      id: 'task-generic',
      name: 'Task',
      icon: Workflow,
      stepType: 'task',
    },
    {
      id: 'control',
      name: 'Control',
      icon: GitBranch,
      stepType: 'control',
    },
  ]

  for (const wt of workTypes.value) {
    const slug = wt.slug.toLowerCase()
    list.push({
      id: `wt-${wt.id}`,
      name: wt.name,
      icon: workTypeIcons[slug] ?? Workflow,
      stepType: 'task',
      workTypeId: wt.id,
    })
  }

  return list
})

function onDragStart(event: DragEvent, item: ToolbarItem) {
  if (!event.dataTransfer) return
  event.dataTransfer.setData('application/cactus-step', JSON.stringify({
    stepType: item.stepType,
    workTypeId: item.workTypeId,
    name: item.name,
  }))
  event.dataTransfer.effectAllowed = 'move'
}

function onDoubleClick(item: ToolbarItem) {
  // Add at canvas center -- approximate position
  emit('addStep', item.stepType, item.workTypeId, { x: 300, y: 200 })
}

onMounted(async () => {
  try {
    workTypes.value = await fetchWorkTypes()
  }
  catch {
    // Non-critical: generic steps still available
  }
})
</script>

<template>
  <div
    class="flex flex-col border-r bg-background transition-all duration-200"
    :class="collapsed ? 'w-12' : 'w-60'"
  >
    <!-- Header -->
    <div class="flex items-center justify-between border-b p-2">
      <span
        v-if="!collapsed"
        class="text-sm font-semibold"
      >
        {{ t('editor.steps') }}
      </span>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="collapsed = !collapsed"
      >
        <PanelLeftClose v-if="!collapsed" class="h-4 w-4" />
        <PanelLeftOpen v-else class="h-4 w-4" />
      </Button>
    </div>

    <!-- Items -->
    <ScrollArea v-if="!collapsed" class="flex-1">
      <div class="space-y-1 p-2">
        <div
          v-for="item in items"
          :key="item.id"
          class="flex cursor-grab items-center gap-2 rounded-md border px-3 py-2 text-sm hover:bg-accent active:cursor-grabbing"
          draggable="true"
          @dragstart="onDragStart($event, item)"
          @dblclick="onDoubleClick(item)"
        >
          <GripVertical class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          <component :is="item.icon" class="h-4 w-4 shrink-0" />
          <span class="truncate">{{ item.name }}</span>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

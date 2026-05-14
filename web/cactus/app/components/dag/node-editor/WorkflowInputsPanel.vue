<script setup lang="ts">
import { MoreHorizontal, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import type { WorkflowInputField } from './workflow-input-utils'
import { workflowInputPath } from './workflow-input-utils'

defineProps<{
  inputs: WorkflowInputField[]
}>()

const emit = defineEmits<{
  insertExpression: [expression: string]
  editInput: [input: WorkflowInputField]
  deleteInput: [input: WorkflowInputField]
}>()

const { t } = useI18n()

function onDragStart(event: DragEvent, input: WorkflowInputField) {
  event.dataTransfer?.setData('application/cactus-expression', workflowInputPath(input.name))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'copy'
}

function onNewDragStart(event: DragEvent) {
  event.dataTransfer?.setData('application/cactus-workflow-input-new', '1')
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'copy'
}
</script>

<template>
  <div class="space-y-1">
    <div
      v-for="input in inputs"
      :key="input.name"
      class="flex cursor-grab items-center gap-2 rounded px-2 py-1 text-sm hover:bg-accent"
      draggable="true"
      @dragstart="onDragStart($event, input)"
      @click="emit('insertExpression', workflowInputPath(input.name))"
    >
      <span class="inline-block h-2 w-2 shrink-0 rounded-full bg-amber-500" />
      <span class="min-w-0 flex-1 truncate font-medium">{{ input.name }}</span>
      <Badge variant="secondary" class="h-5 px-1.5 text-[10px]">{{ input.type }}</Badge>
      <Badge v-if="input.required" variant="outline" class="h-5 px-1.5 text-[10px]">
        {{ t('nodeEditor.required') }}
      </Badge>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            class="ml-auto h-6 w-6 shrink-0"
            :aria-label="t('workflowInputs.actions')"
            :title="t('workflowInputs.actions')"
            :data-testid="`workflow-input-actions-${input.name}`"
            @click.stop
          >
            <MoreHorizontal class="h-3.5 w-3.5" />
            <span class="sr-only">...</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-40" @click.stop>
          <DropdownMenuItem
            class="cursor-pointer gap-2"
            :data-testid="`edit-input-${input.name}`"
            @click.stop="emit('editInput', input)"
          >
            <Pencil class="h-3.5 w-3.5" />
            {{ t('workflowInputs.editInput') }}
          </DropdownMenuItem>
          <DropdownMenuItem
            class="cursor-pointer gap-2 text-destructive focus:text-destructive"
            :data-testid="`delete-input-${input.name}`"
            @click.stop="emit('deleteInput', input)"
          >
            <Trash2 class="h-3.5 w-3.5" />
            {{ t('workflowInputs.deleteInput') }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <div
      class="flex cursor-grab items-center gap-2 rounded border border-dashed px-2 py-1 text-sm text-muted-foreground hover:bg-accent"
      draggable="true"
      @dragstart="onNewDragStart"
    >
      <Plus class="h-3.5 w-3.5" />
      <span>{{ t('nodeEditor.newWorkflowInput') }}</span>
    </div>
  </div>
</template>

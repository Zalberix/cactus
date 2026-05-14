<script setup lang="ts">
import { Braces, Type } from 'lucide-vue-next'
import { Input } from '~/components/ui/input'
import { Button } from '~/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '~/components/ui/tooltip'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  fieldKey: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focus: []
  createWorkflowInput: [fieldKey: string]
}>()

const isExpressionMode = ref(false)

const isExpression = computed(() =>
  isExpressionMode.value || props.modelValue.startsWith('$.'),
)

function toggleMode() {
  isExpressionMode.value = !isExpressionMode.value
}

function onInput(value: string | number) {
  emit('update:modelValue', String(value))
}

function onDrop(event: DragEvent) {
  event.preventDefault()
  const newWorkflowInput = event.dataTransfer?.getData('application/cactus-workflow-input-new')
  if (newWorkflowInput) {
    emit('createWorkflowInput', props.fieldKey)
    return
  }
  const expression = event.dataTransfer?.getData('application/cactus-expression')
  if (!expression) return
  isExpressionMode.value = true
  emit('update:modelValue', expression)
}

// Allow external expression insertion
function insertExpression(expr: string) {
  isExpressionMode.value = true
  emit('update:modelValue', expr)
}

defineExpose({ insertExpression, fieldKey: props.fieldKey })
</script>

<template>
  <div class="flex items-center gap-1">
    <div class="relative flex-1">
      <Input
        :model-value="modelValue"
        :placeholder="placeholder"
        :class="isExpression ? 'font-mono text-xs bg-orange-50 dark:bg-orange-950/30 border-orange-300 dark:border-orange-700' : ''"
        @focus="emit('focus')"
        @dragover.prevent
        @drop="onDrop"
        @update:model-value="onInput"
      />
    </div>
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8 shrink-0"
            :class="isExpression ? 'text-orange-600 dark:text-orange-400' : ''"
            @click="toggleMode"
          >
            <Braces v-if="isExpression" class="h-4 w-4" />
            <Type v-else class="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          {{ isExpression ? 'Switch to fixed value' : 'Switch to expression' }}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  </div>
</template>

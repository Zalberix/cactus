<script setup lang="ts">
import { Label } from '~/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'
import ExpressionField from './ExpressionField.vue'

const props = defineProps<{
  modelValue: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
}>()

const activeField = ref<'left' | 'right'>('left')

const operators = [
  'eq',
  'neq',
  'gt',
  'gte',
  'lt',
  'lte',
  'contains',
  'not_contains',
  'exists',
  'not_exists',
]

const operator = computed(() => stringValue(props.modelValue.operator) || 'eq')
const rightDisabled = computed(() => operator.value === 'exists' || operator.value === 'not_exists')

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function updateField(field: string, value: string) {
  const next = {
    ...props.modelValue,
    [field]: value,
  }
  if ((field === 'operator' && (value === 'exists' || value === 'not_exists')) || rightDisabled.value) {
    next.right = ''
  }
  emit('update:modelValue', next)
}

function insertExpression(expression: string) {
  updateField(activeField.value, expression)
}

defineExpose({ insertExpression })
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-1.5">
      <Label class="text-xs font-medium">Left</Label>
      <ExpressionField
        data-testid="condition-left"
        :model-value="stringValue(modelValue.left)"
        field-key="left"
        placeholder="$.message.value.field"
        @focus="activeField = 'left'"
        @update:model-value="updateField('left', $event)"
      />
    </div>

    <div class="space-y-1.5">
      <Label class="text-xs font-medium">Operator</Label>
      <Select
        :model-value="operator"
        @update:model-value="updateField('operator', String($event))"
      >
        <SelectTrigger data-testid="condition-operator">
          <SelectValue placeholder="Operator" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="item in operators" :key="item" :value="item">
            {{ item }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="!rightDisabled" class="space-y-1.5">
      <Label class="text-xs font-medium">Right</Label>
      <ExpressionField
        data-testid="condition-right"
        :model-value="stringValue(modelValue.right)"
        field-key="right"
        placeholder="$.steps.2.output.value"
        @focus="activeField = 'right'"
        @update:model-value="updateField('right', $event)"
      />
    </div>
  </div>
</template>

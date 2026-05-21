<script setup lang="ts">
import { Plus, Trash2 } from 'lucide-vue-next'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import { createSwitchCaseId, normalizeSwitchCases } from '~/composables/control-outcomes'
import ExpressionField from './ExpressionField.vue'

const props = defineProps<{
  modelValue: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
}>()

const { t } = useI18n()
const localSettings = ref<Record<string, unknown>>({})

watch(
  () => props.modelValue,
  (value) => {
    localSettings.value = {
      ...value,
      cases: normalizeSwitchCases(value.cases),
    }
  },
  { immediate: true },
)

const cases = computed(() => normalizeSwitchCases(localSettings.value.cases))

function expression(): string {
  return typeof localSettings.value.expression === 'string' ? localSettings.value.expression : ''
}

function emitNext(next: Record<string, unknown>) {
  localSettings.value = {
    expression: expression(),
    cases: cases.value,
    ...localSettings.value,
    ...next,
  }
  emit('update:modelValue', { ...localSettings.value })
}

function updateExpression(value: string) {
  emitNext({ expression: value })
}

function addCase() {
  const nextCases = [
    ...cases.value,
    { id: createSwitchCaseId('output'), label: 'Output', value: '' },
  ]
  emitNext({ cases: nextCases })
}

function updateCase(caseID: string, field: 'label' | 'value', value: string) {
  emitNext({
    cases: cases.value.map((item) => {
      if (item.id !== caseID) return item
      return { ...item, [field]: value }
    }),
  })
}

function removeCase(caseID: string) {
  if (typeof window !== 'undefined' && typeof window.confirm === 'function') {
    const ok = window.confirm('Remove this switch output? Existing connections may become invalid.')
    if (!ok) return
  }
  emitNext({ cases: cases.value.filter(item => item.id !== caseID) })
}

function insertExpression(expressionValue: string) {
  updateExpression(expressionValue)
}

defineExpose({ insertExpression })
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-1.5">
      <Label class="text-xs font-medium">Expression</Label>
      <ExpressionField
        data-testid="switch-expression"
        :model-value="expression()"
        field-key="expression"
        placeholder="$.message.value.type"
        @update:model-value="updateExpression"
      />
    </div>

    <div class="space-y-2">
      <div class="flex items-center justify-between gap-2">
        <Label class="text-xs font-medium">{{ t('nodeEditor.switchOutputs') }}</Label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="h-8 gap-1.5"
          data-testid="add-switch-case"
          @click="addCase"
        >
          <Plus class="h-3.5 w-3.5" />
          {{ t('nodeEditor.addSwitchOutput') }}
        </Button>
      </div>

      <div v-for="item in cases" :key="item.id" class="grid grid-cols-[1fr_1fr_auto] items-end gap-2">
        <div class="space-y-1">
          <Label class="text-[11px] text-muted-foreground">{{ t('nodeEditor.outputName') }}</Label>
          <Input
            :data-testid="`switch-case-label-${item.id}`"
            :model-value="item.label"
            @update:model-value="updateCase(item.id, 'label', String($event))"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-[11px] text-muted-foreground">{{ t('nodeEditor.matchValue') }}</Label>
          <Input
            :data-testid="`switch-case-value-${item.id}`"
            :model-value="item.value"
            @update:model-value="updateCase(item.id, 'value', String($event))"
          />
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-9 w-9 text-muted-foreground"
          @click="removeCase(item.id)"
        >
          <Trash2 class="h-4 w-4" />
        </Button>
      </div>

      <div class="rounded border border-dashed px-3 py-2 text-xs text-muted-foreground">
        {{ t('nodeEditor.defaultOutput') }}
      </div>
    </div>
  </div>
</template>

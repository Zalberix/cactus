<script setup lang="ts">
import type { StepData } from '~/composables/useDagEditor'
import { Button } from '~/components/ui/button'
import { ScrollArea } from '~/components/ui/scroll-area'
import DynamicSettingsForm from '~/components/forms/DynamicSettingsForm.vue'
import { useControlSteps } from '~/composables/useControlSteps'
import ConditionControlForm from './ConditionControlForm.vue'
import SwitchControlForm from './SwitchControlForm.vue'

const props = defineProps<{
  stepData: StepData
}>()

const emit = defineEmits<{
  saveControlSettings: [settingsData: Record<string, unknown>]
}>()

const { t } = useI18n()
const controls = useControlSteps()
const formData = ref<Record<string, unknown>>({})
const controlFormRef = ref<InstanceType<typeof ConditionControlForm> | InstanceType<typeof SwitchControlForm> | null>(null)

const controlKind = computed(() => props.stepData.controlKind ?? '')
const definition = computed(() => controls.find(item => item.kind === controlKind.value))
const settingsSchema = computed(() => definition.value?.settingsSchema ?? { type: 'object', properties: {} })

watch(
  () => props.stepData,
  (data) => {
    formData.value = { ...(data.controlSettings ?? {}) }
  },
  { immediate: true },
)

function onSettingsUpdate(next: Record<string, unknown>) {
  formData.value = next
}

function settingsForSave(): Record<string, unknown> {
  const next = { ...formData.value }
  if (controlKind.value === 'condition' && (next.operator === 'exists' || next.operator === 'not_exists')) {
    delete next.right
  }
  return next
}

function onSaveControlSettings() {
  emit('saveControlSettings', settingsForSave())
}

function insertExpression(expression: string) {
  controlFormRef.value?.insertExpression(expression)
}

defineExpose({ insertExpression })
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.parameters') }}</h3>
    </div>

    <ScrollArea class="flex-1 overflow-hidden">
      <div class="p-4">
        <ConditionControlForm
          v-if="controlKind === 'condition'"
          ref="controlFormRef"
          :model-value="formData"
          @update:model-value="onSettingsUpdate"
        />
        <SwitchControlForm
          v-else-if="controlKind === 'switch'"
          ref="controlFormRef"
          :model-value="formData"
          @update:model-value="onSettingsUpdate"
        />
        <DynamicSettingsForm
          v-else-if="controlKind === 'delay'"
          :schema="settingsSchema"
          :model-value="formData"
          @update:model-value="onSettingsUpdate"
        />
        <p v-else class="text-sm text-muted-foreground">{{ t('nodeEditor.noConfiguration') }}</p>
      </div>
    </ScrollArea>

    <div class="border-t p-4 shrink-0">
      <Button
        data-testid="save-control-settings"
        class="w-full"
        @click="onSaveControlSettings"
      >
        {{ t('nodeEditor.saveSettings') }}
      </Button>
    </div>
  </div>
</template>

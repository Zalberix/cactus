<script setup lang="ts">
import type { StepData } from '~/composables/useDagEditor'
import { Button } from '~/components/ui/button'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Separator } from '~/components/ui/separator'
import { Label } from '~/components/ui/label'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '~/components/ui/tabs'
import DynamicSettingsForm from '~/components/forms/DynamicSettingsForm.vue'
import ExpressionField from './ExpressionField.vue'

const props = defineProps<{
  stepData: StepData
}>()

const emit = defineEmits<{
  save: [config: Record<string, unknown>, inputMapping: Record<string, string>]
}>()

const { t } = useI18n()

const controlSchemas: Record<string, Record<string, unknown>> = {
  condition: {
    type: 'object',
    properties: {
      left: { type: 'string', description: 'Value 1', required: true },
      operator: { type: 'string', enum: ['eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'contains', 'not_contains', 'exists', 'not_exists'], required: true },
      right: { type: 'string', description: 'Value 2', required: true },
      combine: { type: 'string', enum: ['AND', 'OR'], default: 'AND' },
    },
  },
  delay: {
    type: 'object',
    properties: {
      duration: { type: 'integer', description: 'Duration', default: 1, required: true },
      unit: { type: 'string', enum: ['seconds', 'minutes', 'hours'], default: 'seconds', required: true },
    },
  },
  switch: {
    type: 'object',
    properties: {
      expression: { type: 'string', description: 'Expression to evaluate' },
      default_case: { type: 'string', description: 'Default case label' },
    },
  },
}

const settingsSchema = computed(() => {
  if (props.stepData.stepType === 'control' && props.stepData.controlKind) {
    return controlSchemas[props.stepData.controlKind] ?? { type: 'object', properties: {} }
  }
  return props.stepData.settingsSchema ?? { type: 'object', properties: {} }
})

const inputFields = computed(() => {
  const schema = props.stepData.inputSchema as { properties?: Record<string, unknown> } | undefined
  if (!schema?.properties) return []
  return Object.keys(schema.properties)
})

const formData = ref<Record<string, unknown>>({})
const mappingData = ref<Record<string, string>>({})

watch(
  () => props.stepData,
  (data) => {
    formData.value = data.stepType === 'control'
      ? { ...(data.controlSettings ?? {}) }
      : { ...(data.config ?? {}) }

    if (Array.isArray(data.inputMapping)) {
      const map: Record<string, string> = {}
      for (const entry of data.inputMapping as Array<{ target?: string; source?: string }>) {
        if (entry.target) map[entry.target] = entry.source ?? ''
      }
      mappingData.value = map
    }
    else {
      mappingData.value = { ...(data.inputMapping ?? {}) }
    }
  },
  { immediate: true },
)

const hasSettings = computed(() => {
  const s = settingsSchema.value as { properties?: Record<string, unknown> }
  return s.properties && Object.keys(s.properties).length > 0
})

const hasMapping = computed(() => inputFields.value.length > 0)

function onSave() {
  emit('save', { ...formData.value }, { ...mappingData.value })
}

function insertExpression(expr: string) {
  // TODO: insert into focused expression field
}

defineExpose({ insertExpression })
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.parameters') || 'Parameters' }}</h3>
    </div>

    <Tabs :default-value="hasSettings ? 'settings' : 'mapping'" class="flex flex-1 flex-col overflow-hidden">
      <TabsList class="mx-4 mt-2 w-auto shrink-0">
        <TabsTrigger v-if="hasSettings" value="settings">
          Settings
        </TabsTrigger>
        <TabsTrigger v-if="hasMapping" value="mapping">
          Input Mapping
        </TabsTrigger>
      </TabsList>

      <!-- Settings tab -->
      <TabsContent v-if="hasSettings" value="settings" class="flex-1 overflow-hidden mt-0">
        <ScrollArea class="h-full">
          <div class="p-4">
            <DynamicSettingsForm
              :schema="settingsSchema"
              :model-value="formData"
              @update:model-value="formData = $event"
            />
          </div>
        </ScrollArea>
      </TabsContent>

      <!-- Input Mapping tab -->
      <TabsContent v-if="hasMapping" value="mapping" class="flex-1 overflow-hidden mt-0">
        <ScrollArea class="h-full">
          <div class="p-4 space-y-3">
            <p class="text-xs text-muted-foreground mb-3">
              Map upstream outputs to this step's input fields.
            </p>
            <div v-for="field in inputFields" :key="field" class="space-y-1">
              <Label class="text-xs font-medium">{{ field }}</Label>
              <ExpressionField
                :model-value="mappingData[field] ?? ''"
                :field-key="field"
                :placeholder="`$.steps.{id}.output.${field}`"
                @update:model-value="mappingData[field] = $event"
              />
            </div>
          </div>
        </ScrollArea>
      </TabsContent>

      <!-- Fallback if neither -->
      <div v-if="!hasSettings && !hasMapping" class="flex-1 flex items-center justify-center">
        <p class="text-sm text-muted-foreground">No configuration available.</p>
      </div>
    </Tabs>

    <div class="border-t p-4 shrink-0">
      <Button class="w-full" @click="onSave">
        {{ t('nodeEditor.save') || 'Save' }}
      </Button>
    </div>
  </div>
</template>

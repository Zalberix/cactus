<script setup lang="ts">
import type { StepData } from '~/composables/useDagEditor'
import { Button } from '~/components/ui/button'
import { ScrollArea } from '~/components/ui/scroll-area'
import { Label } from '~/components/ui/label'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '~/components/ui/tabs'
import DynamicSettingsForm from '~/components/forms/DynamicSettingsForm.vue'
import ExpressionField from './ExpressionField.vue'
import { mappingRecordToEntries, setMappingExpression } from './mapping-utils'

const props = defineProps<{
  stepData: StepData
}>()

const emit = defineEmits<{
  save: [settingsData: Record<string, unknown>, inputMapping: Array<{ target: string, source: string }>]
  createWorkflowInput: []
}>()

const { t } = useI18n()

const settingsSchema = computed(() =>
  props.stepData.settingsSchema ?? { type: 'object', properties: {} },
)

const inputFields = computed(() => {
  const schema = props.stepData.inputSchema as { properties?: Record<string, unknown> } | undefined
  if (!schema?.properties) return []
  return Object.keys(schema.properties)
})

const formData = ref<Record<string, unknown>>({})
const mappingData = ref<Record<string, string>>({})
const activeMappingField = ref<string | null>(null)

watch(
  () => props.stepData,
  (data) => {
    formData.value = { ...(data.config ?? {}) }

    if (Array.isArray(data.inputMapping)) {
      const map: Record<string, string> = {}
      for (const entry of data.inputMapping as Array<{ target?: string, source?: string }>) {
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
const defaultTab = computed(() => (hasMapping.value ? 'mapping' : 'settings'))

function onSave() {
  emit('save', { ...formData.value }, mappingRecordToEntries(mappingData.value))
}

function insertExpression(expr: string) {
  mappingData.value = setMappingExpression(mappingData.value, activeMappingField.value, expr)
}

defineExpose({ insertExpression })
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="border-b px-4 py-3 shrink-0">
      <h3 class="text-sm font-semibold">{{ t('nodeEditor.parameters') }}</h3>
    </div>

    <Tabs :default-value="defaultTab" class="flex flex-1 flex-col overflow-hidden">
      <TabsList class="mx-4 mt-2 w-auto shrink-0">
        <TabsTrigger v-if="hasMapping" value="mapping">
          {{ t('editor.inputMapping') }}
        </TabsTrigger>
        <TabsTrigger v-if="hasSettings" value="settings">
          {{ t('nodeEditor.parameters') }}
        </TabsTrigger>
      </TabsList>

      <TabsContent v-if="hasMapping" value="mapping" class="flex-1 overflow-hidden mt-0">
        <ScrollArea class="h-full">
          <div class="p-4 space-y-3">
            <p class="text-xs text-muted-foreground mb-3">
              {{ t('nodeEditor.mappingDescription') }}
            </p>
            <div v-for="field in inputFields" :key="field" class="space-y-1">
              <Label class="text-xs font-medium">{{ field }}</Label>
              <ExpressionField
                :data-test="`mapping-${field}`"
                :model-value="mappingData[field] ?? ''"
                :field-key="field"
                :placeholder="`$.steps.{id}.output.${field}`"
                @focus="activeMappingField = field"
                @create-workflow-input="activeMappingField = field; emit('createWorkflowInput')"
                @update:model-value="mappingData[field] = $event"
              />
            </div>
          </div>
        </ScrollArea>
      </TabsContent>

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

      <div v-if="!hasSettings && !hasMapping" class="flex-1 flex items-center justify-center">
        <p class="text-sm text-muted-foreground">{{ t('nodeEditor.noConfiguration') }}</p>
      </div>
    </Tabs>

    <div class="border-t p-4 shrink-0">
      <Button class="w-full" @click="onSave">
        {{ t('nodeEditor.save') }}
      </Button>
    </div>
  </div>
</template>

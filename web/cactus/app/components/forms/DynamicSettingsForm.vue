<script setup lang="ts">
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import { Switch } from '~/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

interface JsonSchema {
  type?: string
  properties?: Record<string, JsonSchemaProperty>
  required?: string[]
}

interface JsonSchemaProperty {
  type?: string
  description?: string
  default?: unknown
  enum?: string[]
}

const props = defineProps<{
  schema: Record<string, unknown>
  modelValue: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
}>()

const parsedSchema = computed<JsonSchema>(() => props.schema as JsonSchema)

const properties = computed(() => {
  const p = parsedSchema.value.properties
  if (!p) return []

  return Object.entries(p).map(([key, prop]) => ({
    key,
    ...prop,
    isRequired: parsedSchema.value.required?.includes(key) ?? false,
  }))
})

function formatLabel(key: string): string {
  return key
    .replace(/_/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase())
}

function updateField(key: string, value: unknown) {
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: value,
  })
}

function getFieldValue(key: string): unknown {
  return props.modelValue[key] ?? ''
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-for="prop in properties"
      :key="prop.key"
      class="space-y-2"
    >
      <Label :for="`field-${prop.key}`" class="flex items-center gap-1">
        {{ formatLabel(prop.key) }}
        <span v-if="prop.isRequired" class="text-red-500">*</span>
      </Label>

      <p v-if="prop.description" class="text-xs text-muted-foreground">
        {{ prop.description }}
      </p>

      <!-- Boolean: Switch -->
      <div v-if="prop.type === 'boolean'" class="flex items-center gap-2">
        <Switch
          :id="`field-${prop.key}`"
          :model-value="!!getFieldValue(prop.key)"
          @update:model-value="updateField(prop.key, $event)"
        />
      </div>

      <!-- Enum: Select -->
      <Select
        v-else-if="prop.type === 'string' && prop.enum"
        :model-value="String(getFieldValue(prop.key) || '')"
        @update:model-value="updateField(prop.key, $event)"
      >
        <SelectTrigger :id="`field-${prop.key}`">
          <SelectValue :placeholder="`Select ${formatLabel(prop.key)}`" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="option in prop.enum"
            :key="option"
            :value="option"
          >
            {{ option }}
          </SelectItem>
        </SelectContent>
      </Select>

      <!-- Number / Integer: Input type=number -->
      <Input
        v-else-if="prop.type === 'number' || prop.type === 'integer'"
        :id="`field-${prop.key}`"
        type="number"
        :model-value="getFieldValue(prop.key) as number"
        :required="prop.isRequired"
        @update:model-value="updateField(prop.key, Number($event))"
      />

      <!-- String (default): Input type=text -->
      <Input
        v-else
        :id="`field-${prop.key}`"
        type="text"
        :model-value="String(getFieldValue(prop.key) || '')"
        :required="prop.isRequired"
        @update:model-value="updateField(prop.key, $event)"
      />
    </div>

    <div v-if="properties.length === 0" class="text-sm text-muted-foreground py-4 text-center">
      No settings properties defined in schema.
    </div>
  </div>
</template>

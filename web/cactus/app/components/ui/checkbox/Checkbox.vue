<script setup lang="ts">
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '~/lib/utils'
import { Check, Minus } from 'lucide-vue-next'
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'

const props = defineProps<{
  class?: HTMLAttributes['class']
  disabled?: boolean
  required?: boolean
  name?: string
  value?: string
  id?: string
  checked?: boolean | 'indeterminate'
  defaultChecked?: boolean
}>()

const emits = defineEmits<{
  'update:checked': [value: boolean | 'indeterminate']
}>()

// Map checked/update:checked to v-model (modelValue/update:modelValue) for reka-ui v2
const model = computed({
  get: () => props.checked,
  set: (val) => emits('update:checked', val as boolean | 'indeterminate'),
})
</script>

<template>
  <CheckboxRoot
    v-model="model"
    :disabled="props.disabled"
    :required="props.required"
    :name="props.name"
    :value="props.value"
    :id="props.id"
    :class="
      cn('peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground data-[state=indeterminate]:bg-primary data-[state=indeterminate]:text-primary-foreground',
         props.class)"
  >
    <CheckboxIndicator class="flex h-full w-full items-center justify-center text-current">
      <slot>
        <Minus v-if="model === 'indeterminate'" class="h-4 w-4" />
        <Check v-else class="h-4 w-4" />
      </slot>
    </CheckboxIndicator>
  </CheckboxRoot>
</template>

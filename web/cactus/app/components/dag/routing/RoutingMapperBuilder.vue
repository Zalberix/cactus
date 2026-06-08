<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import {
  builderStateToRulesJson,
  createMapperDefaultRow,
  createMapperMappingRow,
  mapperBuilderWarnings,
  rulesToBuilderState,
  type MapperValueType,
} from './mapper-rules'

const props = defineProps<{
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()
const state = ref(rulesToBuilderState(props.modelValue))
const valueTypes: MapperValueType[] = ['string', 'number', 'boolean', 'null', 'json']

const warnings = computed(() => mapperBuilderWarnings(state.value))

watch(() => props.modelValue, (value) => {
  if (value === builderStateToRulesJson(state.value)) return
  state.value = rulesToBuilderState(value)
})

watch(state, (value) => {
  emit('update:modelValue', builderStateToRulesJson(value))
}, { deep: true })

function addDefault() {
  state.value.defaults.push(createMapperDefaultRow())
}

function removeDefault(index: number) {
  state.value.defaults.splice(index, 1)
}

function addMapping() {
  state.value.mappings.push(createMapperMappingRow())
}

function removeMapping(index: number) {
  state.value.mappings.splice(index, 1)
}
</script>

<template>
  <div class="space-y-4 rounded-md border bg-muted/20 p-3">
    <label class="flex items-center gap-2 text-sm">
      <input
        v-model="state.copyAll"
        type="checkbox"
        class="h-4 w-4 rounded border"
        :disabled="disabled"
      >
      <span>{{ t('workflowRouting.mapperCopyAll') }}</span>
    </label>

    <section class="space-y-2">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h4 class="text-sm font-medium">
            {{ t('workflowRouting.mapperDefaults') }}
          </h4>
          <p class="text-xs text-muted-foreground">
            {{ t('workflowRouting.mapperDefaultsHint') }}
          </p>
        </div>
        <Button type="button" size="sm" variant="outline" :disabled="disabled" @click="addDefault">
          {{ t('workflowRouting.mapperAddDefault') }}
        </Button>
      </div>

      <div v-if="state.defaults.length === 0" class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        {{ t('workflowRouting.mapperNoDefaults') }}
      </div>

      <div
        v-for="(row, index) in state.defaults"
        :key="row.id"
        class="grid gap-2 rounded-md border bg-background p-2 md:grid-cols-[1fr_120px_1fr_auto]"
      >
        <Input v-model="row.key" :placeholder="t('workflowRouting.mapperDefaultKey')" :disabled="disabled" />
        <select v-model="row.valueType" class="rounded-md border bg-background px-3 py-2 text-sm" :disabled="disabled">
          <option v-for="valueType in valueTypes" :key="valueType" :value="valueType">
            {{ valueType }}
          </option>
        </select>
        <Input
          v-if="row.valueType !== 'json'"
          v-model="row.value"
          :placeholder="t('workflowRouting.mapperValue')"
          :disabled="disabled || row.valueType === 'null'"
        />
        <textarea
          v-else
          v-model="row.value"
          class="min-h-10 rounded-md border bg-background px-3 py-2 font-mono text-xs"
          :placeholder="t('workflowRouting.mapperJsonValue')"
          :disabled="disabled"
        />
        <Button type="button" size="sm" variant="ghost" :disabled="disabled" @click="removeDefault(index)">
          {{ t('common.delete') }}
        </Button>
      </div>
    </section>

    <section class="space-y-2">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h4 class="text-sm font-medium">
            {{ t('workflowRouting.mapperMappings') }}
          </h4>
          <p class="text-xs text-muted-foreground">
            {{ t('workflowRouting.mapperMappingsHint') }}
          </p>
        </div>
        <Button type="button" size="sm" variant="outline" :disabled="disabled" @click="addMapping">
          {{ t('workflowRouting.mapperAddMapping') }}
        </Button>
      </div>

      <div v-if="state.mappings.length === 0" class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        {{ t('workflowRouting.mapperNoMappings') }}
      </div>

      <div
        v-for="(row, index) in state.mappings"
        :key="row.id"
        class="grid gap-2 rounded-md border bg-background p-2 md:grid-cols-[1fr_120px_1fr_auto]"
      >
        <Input v-model="row.targetPath" :placeholder="t('workflowRouting.mapperTargetPath')" :disabled="disabled" />
        <select v-model="row.mode" class="rounded-md border bg-background px-3 py-2 text-sm" :disabled="disabled">
          <option value="source">
            {{ t('workflowRouting.mapperModeSource') }}
          </option>
          <option value="literal">
            {{ t('workflowRouting.mapperModeLiteral') }}
          </option>
        </select>

        <Input
          v-if="row.mode === 'source'"
          v-model="row.sourcePath"
          :placeholder="t('workflowRouting.mapperSourcePath')"
          :disabled="disabled"
        />
        <div v-else class="grid gap-2 md:grid-cols-[120px_1fr]">
          <select v-model="row.literalType" class="rounded-md border bg-background px-3 py-2 text-sm" :disabled="disabled">
            <option v-for="valueType in valueTypes" :key="valueType" :value="valueType">
              {{ valueType }}
            </option>
          </select>
          <Input
            v-if="row.literalType !== 'json'"
            v-model="row.literalValue"
            :placeholder="t('workflowRouting.mapperLiteralValue')"
            :disabled="disabled || row.literalType === 'null'"
          />
          <textarea
            v-else
            v-model="row.literalValue"
            class="min-h-10 rounded-md border bg-background px-3 py-2 font-mono text-xs"
            :placeholder="t('workflowRouting.mapperJsonValue')"
            :disabled="disabled"
          />
        </div>

        <Button type="button" size="sm" variant="ghost" :disabled="disabled" @click="removeMapping(index)">
          {{ t('common.delete') }}
        </Button>
      </div>
    </section>

    <div v-if="warnings.length > 0" class="rounded-md border border-amber-300 bg-amber-50 p-3 text-xs text-amber-900">
      <div v-for="warning in warnings" :key="warning">
        {{ warning }}
      </div>
    </div>
  </div>
</template>

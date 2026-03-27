<script setup lang="ts">
import type { Version } from '~/composables/useVersions'
import { Badge } from '~/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

const props = defineProps<{
  versions: Version[]
  modelValue: number | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const selected = computed({
  get: () => props.modelValue ? String(props.modelValue) : undefined,
  set: (val) => emit('update:modelValue', val ? Number(val) : null),
})
</script>

<template>
  <Select v-model="selected">
    <SelectTrigger class="w-[220px]">
      <SelectValue :placeholder="t('editor.versions')" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem
        v-for="version in versions"
        :key="version.id"
        :value="String(version.id)"
      >
        <div class="flex items-center gap-2">
          <span>{{ t('editor.versionNumber', { number: version.version_number }) }}</span>
          <Badge
            v-if="version.is_active"
            variant="default"
            class="h-5 px-1.5 text-[10px]"
          >
            {{ t('editor.activeVersion') }}
          </Badge>
          <Badge
            v-if="version.is_valid"
            variant="secondary"
            class="h-5 px-1.5 text-[10px] bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300"
          >
            {{ t('editor.valid') }}
          </Badge>
          <Badge
            v-else
            variant="secondary"
            class="h-5 px-1.5 text-[10px] bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
          >
            {{ t('editor.invalid') }}
          </Badge>
        </div>
      </SelectItem>
    </SelectContent>
  </Select>
</template>

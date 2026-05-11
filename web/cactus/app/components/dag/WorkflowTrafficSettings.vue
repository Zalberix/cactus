<script setup lang="ts">
import type { VersionSummary } from '~/composables/useVersions'
import type { TrafficVersionSetting } from '~/components/dag/traffic-utils'
import { ArrowRight, Split } from 'lucide-vue-next'
import {
  clampTrafficWeight,
  distributePerVersionTraffic,
  normalizeFixedTrafficInput,
} from '~/components/dag/traffic-utils'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import { Input } from '~/components/ui/input'

const props = defineProps<{
  versions: VersionSummary[]
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [payload: { versions: TrafficVersionSetting[] }]
}>()

const { t } = useI18n()

const settings = ref<TrafficVersionSetting[]>([])

const finalWeights = computed(() => distributePerVersionTraffic(settings.value))
const fixedTotal = computed(() => settings.value.reduce(
  (sum, item) => sum + (item.mode === 'fixed' ? clampTrafficWeight(item.weight ?? 0) : 0),
  0,
))
const shareCount = computed(() => settings.value.filter(item => item.mode === 'share').length)
const shareRemainder = computed(() => Math.max(0, 100 - fixedTotal.value))
const canSave = computed(() => fixedTotal.value <= 100 && (shareCount.value > 0 || fixedTotal.value === 100))

watch(
  () => props.versions,
  (versions) => {
    const existing = new Map(settings.value.map(item => [item.version_id, item]))
    settings.value = versions.map((version) => {
      const current = existing.get(version.id)
      if (current) return current
      return {
        version_id: version.id,
        mode: version.traffic_weight > 0 ? 'fixed' : 'share',
        weight: version.traffic_weight > 0 ? version.traffic_weight : undefined,
      }
    })
  },
  { immediate: true },
)

function versionLabel(version: VersionSummary) {
  return version.name || t('editor.versionNumber', { number: version.version_number })
}

function settingFor(versionId: number) {
  return settings.value.find(item => item.version_id === versionId)
}

function finalWeight(versionId: number) {
  return finalWeights.value.find(item => item.version_id === versionId)?.weight ?? 0
}

function setMode(versionId: number, mode: 'share' | 'fixed') {
  settings.value = settings.value.map((item) => {
    if (item.version_id !== versionId) return item
    if (mode === 'share') return { version_id: versionId, mode }
    return { version_id: versionId, mode, weight: finalWeight(versionId) }
  })
}

function onFixedInput(versionId: number, value: string | number) {
  settings.value = normalizeFixedTrafficInput(settings.value, versionId, Number(value))
}

function onSave() {
  emit('save', {
    versions: settings.value.map(item => ({
      version_id: item.version_id,
      mode: item.mode,
      weight: item.mode === 'fixed' ? clampTrafficWeight(item.weight ?? 0) : undefined,
    })),
  })
}
</script>

<template>
  <Card>
    <CardHeader class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle class="flex items-center gap-2">
            <Split class="h-5 w-5" />
            {{ t('workflowTraffic.settingsTitle') }}
          </CardTitle>
          <CardDescription>{{ t('workflowTraffic.settingsDescription') }}</CardDescription>
        </div>
      </div>
    </CardHeader>

    <CardContent class="space-y-5">
      <div v-if="versions.length === 0" class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
        {{ t('workflowTraffic.noActiveVersions') }}
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="version in versions"
          :key="version.id"
          class="grid gap-3 rounded-lg border bg-background p-4 md:grid-cols-[1fr_auto_auto_auto] md:items-center"
        >
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <p class="truncate font-medium">{{ versionLabel(version) }}</p>
              <Badge variant="secondary">v{{ version.version_number }}</Badge>
              <Badge v-if="version.is_control_group" variant="outline">{{ t('workflowTraffic.controlGroup') }}</Badge>
            </div>
            <p class="mt-1 text-sm text-muted-foreground">
              {{ t('workflowTraffic.currentWeight', { weight: version.traffic_weight ?? 0 }) }}
            </p>
          </div>

          <div class="inline-flex rounded-lg border bg-muted p-1">
            <Button
              type="button"
              size="sm"
              :variant="settingFor(version.id)?.mode === 'share' ? 'default' : 'ghost'"
              @click="setMode(version.id, 'share')"
            >
              {{ t('workflowTraffic.share') }}
            </Button>
            <Button
              type="button"
              size="sm"
              :variant="settingFor(version.id)?.mode === 'fixed' ? 'default' : 'ghost'"
              @click="setMode(version.id, 'fixed')"
            >
              {{ t('workflowTraffic.fixed') }}
            </Button>
          </div>

          <div v-if="settingFor(version.id)?.mode === 'fixed'" class="flex items-center gap-2">
            <Input
              class="w-24 text-right"
              type="number"
              min="0"
              max="100"
              :model-value="settingFor(version.id)?.weight ?? 0"
              @update:model-value="value => onFixedInput(version.id, value)"
            />
            <span class="text-sm text-muted-foreground">%</span>
          </div>
          <div v-else class="text-sm text-muted-foreground">
            {{ t('workflowTraffic.shareMode') }}
          </div>

          <div class="flex items-center gap-2 text-sm">
            <span class="text-muted-foreground">{{ t('workflowTraffic.preview') }}</span>
            <ArrowRight class="h-4 w-4 text-muted-foreground" />
            <span class="font-semibold">{{ finalWeight(version.id) }}%</span>
          </div>
        </div>
      </div>

      <div class="flex flex-col gap-3 border-t pt-4 md:flex-row md:items-center md:justify-between">
        <div class="space-y-1 text-sm">
          <div class="flex flex-wrap gap-x-4 gap-y-1">
            <span>
              <span class="text-muted-foreground">{{ t('workflowTraffic.fixedTotal') }}: </span>
              <span class="font-semibold" :class="{ 'text-destructive': fixedTotal > 100 }">{{ fixedTotal }}%</span>
            </span>
            <span>
              <span class="text-muted-foreground">{{ t('workflowTraffic.sharedRemainder') }}: </span>
              <span class="font-semibold">{{ shareRemainder }}%</span>
            </span>
          </div>
          <p v-if="!canSave" class="text-xs text-destructive">
            {{ t('workflowTraffic.noShareForRemainder') }}
          </p>
        </div>

        <Button :disabled="versions.length === 0 || !canSave || saving" @click="onSave">
          {{ saving ? t('common.loading') : t('workflowTraffic.save') }}
        </Button>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StatusBadge from '~/components/feedback/StatusBadge.vue'
import JsonValueBlock from '~/components/messages/JsonValueBlock.vue'
import RunDuration from '~/components/messages/RunDuration.vue'
import type { GraphStep, MappingEntry, StepRunDetail } from '~/composables/useMessages'

const props = defineProps<{
  step: GraphStep | null
  runtime?: StepRunDetail
  messageValue: Record<string, unknown>
  runSteps: Map<number, StepRunDetail>
}>()

const { t } = useI18n()

type StateTone = 'default' | 'muted' | 'warning' | 'error'

function mapStatus(status?: string): 'pending' | 'running' | 'done' | 'error' | 'working' {
  if (status === 'completed' || status === 'done') return 'done'
  if (status === 'failed' || status === 'error') return 'error'
  if (status === 'skipped') return 'working'
  if (status === 'running') return 'running'
  return 'pending'
}

function inputStateForMapping(mapping: MappingEntry): { value?: unknown; stateText?: string; tone?: StateTone } {
  if (props.runtime?.input_data && mapping.target in props.runtime.input_data) {
    return { value: props.runtime.input_data[mapping.target] }
  }
  if (mapping.source.startsWith('$.message.value.')) {
    const field = mapping.source.replace('$.message.value.', '')
    return field in props.messageValue
      ? { value: props.messageValue[field] }
      : { stateText: t('messages.detail.messageFieldAbsent', { field }), tone: 'warning' }
  }
  if (mapping.source.startsWith('$.steps.')) {
    const match = mapping.source.match(/^\$\.steps\.(\d+)\.output\.(.+)$/)
    const sourceStepId = match ? Number(match[1]) : 0
    const sourceRuntime = props.runSteps.get(sourceStepId)
    if (!sourceRuntime || sourceRuntime.status === 'pending' || sourceRuntime.status === 'running') {
      return { stateText: t('messages.detail.waitingForStep', { id: sourceStepId }), tone: 'muted' }
    }
    if (sourceRuntime.status === 'failed' || sourceRuntime.status === 'skipped') {
      return {
        stateText: t('messages.detail.sourceStepStatus', { id: sourceStepId, status: sourceRuntime.status }),
        tone: 'warning',
      }
    }
    return {
      stateText: t('messages.detail.outputFieldMissing', { id: sourceStepId }),
      tone: 'warning',
    }
  }
  return { value: mapping.source }
}

const title = computed(() => {
  if (!props.step) return t('messages.detail.noStepSelected')
  if (props.step.control_kind === 'start' && !props.step.name) return 'System Trigger'
  return props.step.name ?? props.step.work_type_name ?? props.step.control_kind ?? `Step ${props.step.id}`
})

const outputState = computed<{ value?: unknown; stateText?: string; tone?: StateTone }>(() => {
  if (props.runtime?.output_data !== undefined) return { value: props.runtime.output_data }
  switch (props.runtime?.status) {
    case 'running':
      return { stateText: t('messages.detail.outputRunning'), tone: 'muted' }
    case 'skipped':
      return { stateText: t('messages.detail.outputSkipped'), tone: 'warning' }
    case 'failed':
      return { stateText: props.runtime.error_message ?? t('messages.detail.workflowFailed'), tone: 'error' }
    default:
      return { stateText: t('messages.detail.outputPending'), tone: 'muted' }
  }
})
</script>

<template>
  <aside class="min-h-0 overflow-auto border-l bg-background p-4">
    <div v-if="!step" class="flex h-full items-center justify-center text-sm text-muted-foreground">
      {{ t('messages.detail.noStepSelected') }}
    </div>

    <div v-else class="space-y-5">
      <div class="space-y-2">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h2 class="truncate text-base font-semibold">{{ title }}</h2>
            <p class="text-xs text-muted-foreground">#{{ step.id }} - {{ step.step_type }}</p>
          </div>
          <StatusBadge :status="mapStatus(runtime?.status)" />
        </div>
      </div>

      <section class="grid grid-cols-2 gap-3 rounded-md border p-3 text-sm">
        <div>
          <div class="text-xs text-muted-foreground">{{ t('messages.detail.startedAt') }}</div>
          <div>{{ runtime?.started_at ? new Date(runtime.started_at).toLocaleString() : '-' }}</div>
        </div>
        <div>
          <div class="text-xs text-muted-foreground">{{ t('messages.detail.finishedAt') }}</div>
          <div>{{ runtime?.completed_at ? new Date(runtime.completed_at).toLocaleString() : '-' }}</div>
        </div>
        <div class="col-span-2">
          <div class="text-xs text-muted-foreground">{{ t('messages.detail.duration') }}</div>
          <RunDuration
            :started-at="runtime?.started_at"
            :completed-at="runtime?.completed_at"
            :duration-ms="runtime?.duration_ms"
          />
        </div>
      </section>

      <section class="space-y-3">
        <h3 class="text-sm font-semibold">{{ t('messages.detail.input') }}</h3>
        <JsonValueBlock
          v-for="mapping in step.input_mapping"
          :key="mapping.target"
          :title="mapping.target"
          :value="inputStateForMapping(mapping).value"
          :state-text="inputStateForMapping(mapping).stateText"
          :tone="inputStateForMapping(mapping).tone"
        />
        <JsonValueBlock
          v-if="step.input_mapping.length === 0"
          :title="t('messages.detail.input')"
          :state-text="t('messages.detail.noInputMapping')"
          tone="muted"
        />
      </section>

      <JsonValueBlock
        :title="t('messages.detail.output')"
        :value="outputState.value"
        :state-text="outputState.stateText"
        :tone="outputState.tone"
      />

      <JsonValueBlock
        v-if="runtime?.error_message"
        :title="t('messages.detail.error')"
        :state-text="runtime.error_message"
        tone="error"
      />
    </div>
  </aside>
</template>

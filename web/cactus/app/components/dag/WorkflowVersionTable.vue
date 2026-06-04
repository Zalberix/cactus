<script setup lang="ts">
import { FileJson, MoreHorizontal, Pause, Play, Trash2 } from 'lucide-vue-next'
import type { VersionSummary } from '~/composables/useVersions'
import type { SupportedSchemaLabelsByVersionId } from '~/components/dag/supported-schema-utils'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'

const props = defineProps<{
  versions: VersionSummary[]
  orgId: number
  workflowId: number
  supportedSchemaLabelsByVersionId?: SupportedSchemaLabelsByVersionId
}>()

const emit = defineEmits<{
  activate: [version: VersionSummary]
  deactivate: [version: VersionSummary]
  delete: [version: VersionSummary]
  showInputSchema: []
}>()
const { t } = useI18n()

function editPath(version: VersionSummary) {
  return workflowVersionEditorPath(props.orgId, props.workflowId, version.id)
}

function supportedSchemaLabels(version: VersionSummary) {
  return props.supportedSchemaLabelsByVersionId?.[version.id] ?? []
}
</script>

<template>
  <div class="overflow-hidden rounded-md border">
    <table class="w-full text-sm">
      <thead class="border-b bg-muted/40 text-left">
        <tr>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.name') }}</th>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.status') }}</th>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.editing') }}</th>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.supportedSchemas') }}</th>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.runs') }}</th>
          <th class="px-4 py-3 font-medium">{{ t('workflowVersions.headers.traffic') }}</th>
          <th class="w-12 px-4 py-3" />
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="version in versions"
          :key="version.id"
          class="border-b last:border-0"
        >
          <td class="px-4 py-3">
            <div class="flex items-center gap-2">
              <NuxtLink :to="editPath(version)" class="font-medium hover:underline">
                {{ version.name || t('workflowVersions.versionFallback') }}
              </NuxtLink>
              <span class="text-xs text-muted-foreground">v{{ version.version_number }}</span>
            </div>
          </td>
          <td class="px-4 py-3">
            <div class="flex flex-wrap gap-2">
              <Badge v-if="version.is_active" variant="default">{{ t('status.active') }}</Badge>
              <Badge v-else variant="secondary">{{ t('status.inactive') }}</Badge>
              <Badge
                v-if="version.is_active"
                :variant="version.is_valid ? 'secondary' : 'destructive'"
              >
                {{ version.is_valid ? t('editor.valid') : t('workflowVersions.status.validationRequired') }}
              </Badge>
              <Badge v-if="version.is_control_group" variant="outline">{{ t('workflowVersions.status.control') }}</Badge>
            </div>
          </td>
          <td class="px-4 py-3">
            <span class="text-muted-foreground">
              {{ version.is_active && version.run_count > 0 ? t('workflowVersions.editing.readOnly') : t('workflowVersions.editing.editable') }}
            </span>
          </td>
          <td class="px-4 py-3">
            <div v-if="supportedSchemaLabels(version).length > 0" class="flex flex-wrap gap-1.5">
              <Badge
                v-for="schemaLabel in supportedSchemaLabels(version)"
                :key="schemaLabel"
                variant="outline"
              >
                {{ schemaLabel }}
              </Badge>
            </div>
            <span v-else class="text-muted-foreground">-</span>
          </td>
          <td class="px-4 py-3">{{ version.run_count }}</td>
          <td class="px-4 py-3">{{ version.traffic_weight }}%</td>
          <td class="px-4 py-3 text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" class="h-8 w-8">
                  <MoreHorizontal class="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="emit('showInputSchema')">
                  <FileJson class="mr-2 h-4 w-4" />
                  {{ t('workflowVersions.actions.workflowInputSchema') }}
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="!version.is_active"
                  :disabled="!version.is_valid"
                  @click="emit('activate', version)"
                >
                  <Play class="mr-2 h-4 w-4" />
                  {{ t('workflowVersions.actions.activate') }}
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-else
                  @click="emit('deactivate', version)"
                >
                  <Pause class="mr-2 h-4 w-4" />
                  {{ t('workflowVersions.actions.deactivate') }}
                </DropdownMenuItem>
                <DropdownMenuItem
                  class="text-destructive focus:text-destructive"
                  @click="emit('delete', version)"
                >
                  <Trash2 class="mr-2 h-4 w-4" />
                  {{ t('workflowVersions.actions.delete') }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

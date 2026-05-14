<script setup lang="ts">
import { ExternalLink, FileJson, MoreHorizontal, Play, Pause, Trash2 } from 'lucide-vue-next'
import type { VersionSummary } from '~/composables/useVersions'
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
}>()

const emit = defineEmits<{
  activate: [version: VersionSummary]
  deactivate: [version: VersionSummary]
  delete: [version: VersionSummary]
  showInputSchema: []
}>()

function editPath(version: VersionSummary) {
  return workflowVersionEditorPath(props.orgId, props.workflowId, version.id)
}

function editingStatus(version: VersionSummary) {
  if (version.deleted_at) return 'Deleted'
  if (version.is_active && version.run_count > 0) return 'Read-only'
  return 'Editable'
}
</script>

<template>
  <div class="overflow-hidden rounded-md border">
    <table class="w-full text-sm">
      <thead class="border-b bg-muted/40 text-left">
        <tr>
          <th class="px-4 py-3 font-medium">Name</th>
          <th class="px-4 py-3 font-medium">Status</th>
          <th class="px-4 py-3 font-medium">Editing</th>
          <th class="px-4 py-3 font-medium">Runs</th>
          <th class="px-4 py-3 font-medium">Traffic</th>
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
            <div class="font-medium">{{ version.name || `Version ${version.version_number}` }}</div>
            <div class="text-xs text-muted-foreground">v{{ version.version_number }}</div>
          </td>
          <td class="px-4 py-3">
            <div class="flex flex-wrap gap-2">
              <Badge v-if="version.is_active" variant="default">Active</Badge>
              <Badge v-else variant="secondary">Inactive</Badge>
              <Badge v-if="version.deleted_at" variant="outline">Deleted</Badge>
              <Badge :variant="version.is_valid ? 'secondary' : 'destructive'">
                {{ version.is_valid ? 'Valid' : 'Validation required' }}
              </Badge>
              <Badge v-if="version.is_control_group" variant="outline">Control</Badge>
            </div>
          </td>
          <td class="px-4 py-3">
            <span class="text-muted-foreground">{{ editingStatus(version) }}</span>
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
                <DropdownMenuItem
                  :disabled="Boolean(version.deleted_at)"
                  as-child
                >
                  <NuxtLink :to="editPath(version)">
                    <ExternalLink class="mr-2 h-4 w-4" />
                    Open DAG
                  </NuxtLink>
                </DropdownMenuItem>
                <DropdownMenuItem @click="emit('showInputSchema')">
                  <FileJson class="mr-2 h-4 w-4" />
                  Workflow input schema
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="!version.is_active"
                  :disabled="Boolean(version.deleted_at) || !version.is_valid"
                  @click="emit('activate', version)"
                >
                  <Play class="mr-2 h-4 w-4" />
                  Activate
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-else
                  :disabled="Boolean(version.deleted_at)"
                  @click="emit('deactivate', version)"
                >
                  <Pause class="mr-2 h-4 w-4" />
                  Deactivate
                </DropdownMenuItem>
                <DropdownMenuItem
                  class="text-destructive focus:text-destructive"
                  :disabled="Boolean(version.deleted_at)"
                  @click="emit('delete', version)"
                >
                  <Trash2 class="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

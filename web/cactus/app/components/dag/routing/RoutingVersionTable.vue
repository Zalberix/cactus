<script setup lang="ts">
import type { RoutingVersionRowRecord } from '~/composables/useWorkflowRouting'
import { MoreHorizontal } from 'lucide-vue-next'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'

defineProps<{
  rows: RoutingVersionRowRecord[]
  saving?: boolean
}>()

const emit = defineEmits<{
  edit: [row: RoutingVersionRowRecord]
  compatibility: [row: RoutingVersionRowRecord]
  archive: [row: RoutingVersionRowRecord]
}>()

const { t } = useI18n()

function schemaLabel(schema: RoutingVersionRowRecord['supported_schemas'][number]) {
  const support = schema.support_mode === 'mapper'
    ? schema.mapper_id ? `Mapper #${schema.mapper_id}` : t('workflowRouting.compatibility')
    : t('workflowRouting.native')
  return `${schema.schema_code} v${schema.schema_version_number} - ${support}`
}
</script>

<template>
  <div class="overflow-hidden rounded-2xl border bg-background">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{{ t('workflowRouting.processVersion') }}</TableHead>
          <TableHead>{{ t('workflowRouting.supportedSchemaVersions') }}</TableHead>
          <TableHead>{{ t('workflowRouting.activeTesting') }}</TableHead>
          <TableHead class="w-12 text-right">{{ t('common.actions') }}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="rows.length === 0" :colspan="4">
          {{ t('workflowRouting.noRoutingVersions') }}
        </TableEmpty>
        <TableRow v-for="row in rows" :key="row.workflow_version_id" :data-testid="`routing-version-row-${row.workflow_version_id}`">
          <TableCell>
            <div class="font-medium">{{ row.workflow_version_name }}</div>
            <div class="text-xs text-muted-foreground">v{{ row.workflow_version_number }}</div>
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap gap-1.5">
              <Badge
                v-for="schema in row.supported_schemas"
                :key="schema.compatibility_id"
                variant="outline"
              >
                {{ schemaLabel(schema) }}
              </Badge>
            </div>
          </TableCell>
          <TableCell>
            <div v-if="row.active_tests.length > 0" class="flex flex-wrap gap-1.5">
              <Badge v-for="test in row.active_tests" :key="test.id" variant="secondary">
                {{ test.name }}
              </Badge>
            </div>
            <span v-else class="text-sm text-muted-foreground">-</span>
          </TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" class="h-8 w-8" :disabled="saving" :aria-label="t('common.actions')">
                  <MoreHorizontal class="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="emit('edit', row)">
                  {{ t('common.edit') }}
                </DropdownMenuItem>
                <DropdownMenuItem @click="emit('compatibility', row)">
                  {{ t('workflowRouting.compatibility') }}
                </DropdownMenuItem>
                <DropdownMenuItem class="text-destructive focus:text-destructive" @click="emit('archive', row)">
                  {{ t('common.delete') }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>

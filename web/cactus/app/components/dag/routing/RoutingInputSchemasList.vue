<script setup lang="ts">
import type { WorkflowInputSchemaRecord } from '~/composables/useWorkflowRouting'
import { Layers3, Plus } from 'lucide-vue-next'
import { workflowInputFieldsFromSchema } from '~/components/dag/node-editor/workflow-input-utils'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'

defineProps<{
  schemas: WorkflowInputSchemaRecord[]
  saving?: boolean
}>()

const emit = defineEmits<{
  create: []
  edit: [schema: WorkflowInputSchemaRecord]
  setDefault: [schema: WorkflowInputSchemaRecord]
  activate: [schema: WorkflowInputSchemaRecord]
  delete: [schema: WorkflowInputSchemaRecord]
}>()

const { t } = useI18n()

function schemaFields(schema: WorkflowInputSchemaRecord) {
  return workflowInputFieldsFromSchema(schema.schema_json)
}
</script>

<template>
  <Card>
    <CardHeader class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle class="flex items-center gap-2">
            <Layers3 class="h-5 w-5" />
            {{ t('workflowRouting.inputSchemas') }}
          </CardTitle>
          <CardDescription>{{ t('workflowRouting.inputSchemasDescription') }}</CardDescription>
        </div>
        <Button
          type="button"
          class="gap-2"
          :disabled="saving"
          data-testid="open-create-schema-modal"
          @click="emit('create')"
        >
          <Plus class="h-4 w-4" />
          {{ t('workflowRouting.createInputSchema') }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-3">
      <div v-for="schema in schemas" :key="schema.id" class="rounded-2xl border p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ schema.code }} v{{ schema.version_number }}</span>
              <Badge v-if="schema.is_default" variant="default">{{ t('workflowRouting.default') }}</Badge>
              <Badge variant="outline">{{ schema.status }}</Badge>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">Schema #{{ schema.id }}</p>
          </div>
          <div class="flex gap-2">
            <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-schema-${schema.id}`" @click="emit('edit', schema)">
              {{ t('common.edit') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="schema.is_default || schema.status !== 'active' || saving" @click="emit('setDefault', schema)">
              {{ t('workflowRouting.setDefault') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="schema.status === 'active' || saving" @click="emit('activate', schema)">
              {{ t('workflowRouting.activate') }}
            </Button>
            <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-schema-${schema.id}`" @click="emit('delete', schema)">
              {{ t('common.delete') }}
            </Button>
          </div>
        </div>
        <div class="mt-3" :data-testid="`routing-input-schema-field-list-${schema.id}`">
          <div v-if="schemaFields(schema).length === 0" class="text-sm text-muted-foreground">
            {{ t('workflowInputs.empty') }}
          </div>
          <div v-else class="flex flex-wrap gap-2">
            <Badge
              v-for="field in schemaFields(schema)"
              :key="field.name"
              variant="secondary"
              class="gap-1"
              :data-testid="`routing-input-schema-field-${schema.id}-${field.name}`"
            >
              {{ field.name }}
            </Badge>
          </div>
        </div>
      </div>
      <div v-if="schemas.length === 0" class="rounded-2xl border border-dashed p-6 text-sm text-muted-foreground">
        {{ t('workflowRouting.noInputSchemas') }}
      </div>
    </CardContent>
  </Card>
</template>

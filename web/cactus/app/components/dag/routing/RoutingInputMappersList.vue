<script setup lang="ts">
import type { WorkflowInputMapperRecord } from '~/composables/useWorkflowRouting'
import { Plus } from 'lucide-vue-next'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'

defineProps<{
  mappers: WorkflowInputMapperRecord[]
  saving?: boolean
}>()

const emit = defineEmits<{
  create: []
  edit: [mapper: WorkflowInputMapperRecord]
  toggle: [mapper: WorkflowInputMapperRecord]
  delete: [mapper: WorkflowInputMapperRecord]
}>()

const { t } = useI18n()
</script>

<template>
  <Card>
    <CardHeader class="space-y-3">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle>{{ t('workflowRouting.inputMappers') }}</CardTitle>
          <CardDescription>{{ t('workflowRouting.inputMappersDescription') }}</CardDescription>
        </div>
        <Button type="button" class="gap-2" :disabled="saving" data-testid="open-create-mapper-modal" @click="emit('create')">
          <Plus class="h-4 w-4" />
          {{ t('workflowRouting.createMapper') }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-3">
      <div v-for="mapper in mappers" :key="mapper.id" class="rounded-2xl border p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ mapper.name }}</span>
              <Badge variant="outline">{{ mapper.mapper_type }}</Badge>
              <Badge :variant="mapper.is_active ? 'default' : 'outline'">{{ mapper.is_active ? 'active' : 'inactive' }}</Badge>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">Mapper #{{ mapper.id }}</p>
          </div>
          <div class="flex gap-2">
            <Button size="sm" variant="outline" :disabled="saving" :data-testid="`edit-mapper-${mapper.id}`" @click="emit('edit', mapper)">
              {{ t('common.edit') }}
            </Button>
            <Button size="sm" variant="outline" :disabled="saving" @click="emit('toggle', mapper)">
              {{ mapper.is_active ? t('workflowRouting.deactivate') : t('workflowRouting.activate') }}
            </Button>
            <Button size="sm" variant="destructive" :disabled="saving" :data-testid="`delete-mapper-${mapper.id}`" @click="emit('delete', mapper)">
              {{ t('common.delete') }}
            </Button>
          </div>
        </div>
        <pre class="mt-3 max-h-32 overflow-auto rounded-xl bg-muted p-3 text-xs">{{ JSON.stringify(mapper.rules, null, 2) }}</pre>
      </div>
      <div v-if="mappers.length === 0" class="rounded-2xl border border-dashed p-6 text-sm text-muted-foreground">
        {{ t('workflowRouting.noInputMappers') }}
      </div>
    </CardContent>
  </Card>
</template>

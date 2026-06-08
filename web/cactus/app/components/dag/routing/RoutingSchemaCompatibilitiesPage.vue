<script setup lang="ts">
import type { InputSchemaCompatibilityRow, WorkflowInputSchemaRecord } from '~/composables/useWorkflowRouting'
import { ArrowLeft, Link2, Plus } from 'lucide-vue-next'
import {
  workflowInputSchemaCompatibilityCreatePath,
  workflowRoutingPath,
} from '~/composables/useWorkflowRouting'
import { Badge } from '~/components/ui/badge'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const routing = useWorkflowRouting()

const orgId = computed(() => Number(route.params.orgId))
const workflowId = computed(() => Number(route.params.workflowId))
const inputSchemaId = computed(() => Number(route.params.inputSchemaId))

const sourceSchema = ref<WorkflowInputSchemaRecord | null>(null)
const rows = ref<InputSchemaCompatibilityRow[]>([])
const loading = ref(true)

const sourceLabel = computed(() => sourceSchema.value
  ? `${sourceSchema.value.code} v${sourceSchema.value.version_number}`
  : t('common.loading'))

async function loadData() {
  loading.value = true
  try {
    const [schema, compatibilityRows] = await Promise.all([
      routing.fetchInputSchema(inputSchemaId.value),
      routing.fetchInputSchemaCompatibilities(inputSchemaId.value),
    ])
    sourceSchema.value = schema
    rows.value = compatibilityRows
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

function createPath() {
  return workflowInputSchemaCompatibilityCreatePath(orgId.value, workflowId.value, inputSchemaId.value)
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <Button variant="ghost" class="-ml-3 mb-3 gap-2" @click="router.push(workflowRoutingPath(orgId, workflowId))">
          <ArrowLeft class="h-4 w-4" />
          {{ t('common.back') }}
        </Button>
        <div class="flex items-center gap-3">
          <div class="rounded-md bg-primary/10 p-3 text-primary">
            <Link2 class="h-6 w-6" />
          </div>
          <div>
            <h1 class="text-2xl font-semibold">{{ t('workflowRouting.compatibility') }}</h1>
            <p class="text-sm text-muted-foreground">{{ sourceLabel }}</p>
          </div>
        </div>
      </div>
      <Button type="button" class="gap-2" @click="router.push(createPath())">
        <Plus class="h-4 w-4" />
        {{ t('workflowRouting.createCompatibility') }}
      </Button>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>{{ t('workflowRouting.compatibilities') }}</CardTitle>
        <CardDescription>{{ t('workflowRouting.compatibilitiesDescription') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">
          {{ t('common.loading') }}
        </div>
        <div v-else class="overflow-hidden rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t('workflowRouting.processVersion') }}</TableHead>
                <TableHead>{{ t('workflowRouting.fieldWorkflowVersion') }}</TableHead>
                <TableHead>{{ t('workflowRouting.compatibility') }}</TableHead>
                <TableHead>{{ t('workflowRouting.defaultRoute') }}</TableHead>
                <TableHead>{{ t('status.active') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="rows.length === 0" :colspan="5">
                {{ t('workflowRouting.noCompatibilities') }}
              </TableEmpty>
              <TableRow v-for="row in rows" :key="row.id">
                <TableCell class="font-medium">{{ row.workflow_version_name }}</TableCell>
                <TableCell>v{{ row.workflow_version_number }}</TableCell>
                <TableCell>
                  <Badge variant="outline">
                    {{ row.workflow_input_mapper_id ? `Mapper #${row.workflow_input_mapper_id}` : (row.compatibility_type === 'native' ? t('workflowRouting.native') : row.compatibility_type) }}
                  </Badge>
                </TableCell>
                <TableCell>{{ row.is_default_route ? t('workflowRouting.default') : '-' }}</TableCell>
                <TableCell>{{ row.is_active ? t('status.active') : t('status.inactive') }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ArrowLeft } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { System } from '~/composables/useSystems'
import { Button } from '~/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '~/components/ui/form'
import { Input } from '~/components/ui/input'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))
const systemId = computed(() => Number(route.params.systemId))

const { fetchSystems, updateSystem } = useSystems()

const system = ref<System | null>(null)
const loading = ref(true)
const saving = ref(false)

const systemSchema = toTypedSchema(
  z.object({
    name: z.string().min(2, t('systems.validation.nameMin')),
    description: z.string().optional(),
  }),
)

const initialValues = computed(() => ({
  name: system.value?.name ?? '',
  description: system.value?.description ?? '',
}))

async function loadData() {
  loading.value = true
  try {
    const systems = await fetchSystems(orgId.value)
    system.value = systems.find(s => s.id === systemId.value) ?? null
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function onSave(values: Record<string, unknown>) {
  saving.value = true
  try {
    await updateSystem(systemId.value, values.name as string, values.description as string)
    toast({ title: t('systems.updated') })
  }
  catch {
    toast({ title: t('error.server'), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function goBack() {
  router.push(`/org/${orgId.value}/settings/systems`)
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6">
    <Button variant="ghost" size="sm" class="gap-2" @click="goBack">
      <ArrowLeft class="h-4 w-4" />
      {{ t('systems.backToSystems') }}
    </Button>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <span class="text-muted-foreground">{{ t('common.loading') }}</span>
    </div>

    <template v-else-if="system">
      <Card>
        <CardHeader>
          <CardTitle>{{ t('systems.editSystem') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <Form
            v-slot="{ handleSubmit }"
            :validation-schema="systemSchema"
            :initial-values="initialValues"
            class="space-y-4"
          >
            <form @submit="handleSubmit($event, onSave)">
              <div class="space-y-4">
                <FormField v-slot="{ componentField }" name="name">
                  <FormItem>
                    <FormLabel>{{ t('systems.name') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>

                <FormField v-slot="{ componentField }" name="description">
                  <FormItem>
                    <FormLabel>{{ t('systems.description') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>

              <div class="mt-4 flex justify-end">
                <Button type="submit" :disabled="saving">
                  {{ t('common.save') }}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </template>

    <div v-else class="py-12 text-center text-muted-foreground">
      {{ t('error.notFound') }}
    </div>
  </div>
</template>

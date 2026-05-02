<script setup lang="ts">
import { ArrowLeft } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { Permission } from '~/utils/permissions'
import type { Role } from '~/composables/useRoles'
import PermissionGroupEditor from '~/components/forms/PermissionGroupEditor.vue'
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
import { Separator } from '~/components/ui/separator'
import { toast } from '~/components/ui/toast/use-toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgId = computed(() => Number(route.params.orgId))
const roleId = computed(() => Number(route.params.roleId))

const { fetchRoles, updateRole, fetchRolePermissions } = useRoles()

const role = ref<Role | null>(null)
const permissions = ref<Permission[]>([])
const loading = ref(true)
const saving = ref(false)

const roleSchema = toTypedSchema(
  z.object({
    name: z.string().min(2, t('roles.validation.nameMin')),
    description: z.string().optional(),
  }),
)

const initialValues = computed(() => ({
  name: role.value?.name ?? '',
  description: role.value?.description ?? '',
}))

async function loadRole() {
  loading.value = true
  try {
    const roles = await fetchRoles(orgId.value)
    role.value = roles.find(r => r.id === roleId.value) ?? null

    if (role.value) {
      try {
        const perms = await fetchRolePermissions(roleId.value)
        permissions.value = perms.map((p: any) => typeof p === 'string' ? p : p.slug) as Permission[]
      }
      catch {
        // Fallback to role's inline permissions
        const rp = role.value.permissions ?? []
        permissions.value = rp.map((p: any) => typeof p === 'string' ? p : p.slug) as Permission[]
      }
    }
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    loading.value = false
  }
}

async function onSave(values: Record<string, unknown>) {
  saving.value = true
  try {
    await updateRole(roleId.value, {
      name: values.name as string,
      description: (values.description as string) || undefined,
      permissions: permissions.value,
    })
    toast({ title: t('roles.updated') })
  }
  catch (err) {
    toast({ title: getErrorMessage(err, t('error.server')), variant: 'destructive' })
  }
  finally {
    saving.value = false
  }
}

function goBack() {
  router.push(`/org/${orgId.value}/settings/roles`)
}

onMounted(() => {
  loadRole()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Back link -->
    <Button variant="ghost" size="sm" class="gap-2" @click="goBack">
      <ArrowLeft class="h-4 w-4" />
      {{ t('roles.backToRoles') }}
    </Button>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <span class="text-muted-foreground">{{ t('common.loading') }}</span>
    </div>

    <template v-else-if="role">
      <!-- Role Form -->
      <Card>
        <CardHeader>
          <CardTitle>{{ t('roles.editRole') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <Form
            v-slot="{ handleSubmit }"
            :validation-schema="roleSchema"
            :initial-values="initialValues"
            class="space-y-4"
          >
            <form @submit="handleSubmit($event, onSave)">
              <div class="space-y-4">
                <FormField v-slot="{ componentField }" name="name">
                  <FormItem>
                    <FormLabel>{{ t('roles.name') }}</FormLabel>
                    <FormControl>
                      <Input
                        :placeholder="t('roles.namePlaceholder')"
                        v-bind="componentField"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>

                <FormField v-slot="{ componentField }" name="description">
                  <FormItem>
                    <FormLabel>{{ t('roles.description') }}</FormLabel>
                    <FormControl>
                      <Input
                        :placeholder="t('roles.descriptionPlaceholder')"
                        v-bind="componentField"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>

              <Separator class="my-6" />

              <!-- Permissions Editor -->
              <div class="space-y-4">
                <h3 class="text-lg font-medium">{{ t('roles.permissions') }}</h3>
                <PermissionGroupEditor v-model="permissions" />
              </div>

              <div class="mt-6 flex justify-end">
                <Button type="submit" :disabled="saving">
                  {{ t('roles.saveChanges') }}
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

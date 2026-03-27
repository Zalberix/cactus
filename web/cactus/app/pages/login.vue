<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import { useForm } from 'vee-validate'
import * as z from 'zod'

definePageMeta({
  layout: 'auth',
})

const { t } = useI18n()
const authStore = useAuthStore()

const schema = toTypedSchema(
  z.object({
    email: z.string().email(),
    password: z.string().min(6),
  }),
)

const { handleSubmit, errors, defineField, meta } = useForm({
  validationSchema: schema,
})

const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')

const isLoading = ref(false)
const loginError = ref<string | null>(null)

const onSubmit = handleSubmit(async (values) => {
  isLoading.value = true
  loginError.value = null

  try {
    await authStore.login(values.email, values.password)
    await navigateTo('/')
  }
  catch (err: unknown) {
    loginError.value = err instanceof Error ? err.message : String(err)
  }
  finally {
    isLoading.value = false
  }
})
</script>

<template>
  <Card class="w-full max-w-[400px] p-8">
    <CardHeader class="space-y-1 p-0 pb-6">
      <CardTitle class="text-2xl font-semibold">
        {{ t('auth.signIn') }}
      </CardTitle>
    </CardHeader>
    <CardContent class="p-0">
      <form class="space-y-4" @submit="onSubmit">
        <div
          v-if="loginError"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ loginError }}
        </div>

        <FormField v-slot="{ componentField }" name="email">
          <FormItem>
            <FormLabel>{{ t('auth.email') }}</FormLabel>
            <FormControl>
              <Input
                v-model="email"
                type="email"
                :placeholder="t('auth.emailPlaceholder')"
                v-bind="{ ...emailAttrs, ...componentField }"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="password">
          <FormItem>
            <FormLabel>{{ t('auth.password') }}</FormLabel>
            <FormControl>
              <Input
                v-model="password"
                type="password"
                :placeholder="t('auth.passwordPlaceholder')"
                v-bind="{ ...passwordAttrs, ...componentField }"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <Button
          type="submit"
          class="w-full"
          :disabled="isLoading"
        >
          <template v-if="isLoading">
            <svg
              class="mr-2 h-4 w-4 animate-spin"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            {{ t('common.loading') }}
          </template>
          <template v-else>
            {{ t('auth.signIn') }}
          </template>
        </Button>
      </form>
    </CardContent>
  </Card>
</template>

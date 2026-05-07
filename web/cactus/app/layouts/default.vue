<script setup lang="ts">
import { LogOut } from 'lucide-vue-next'
import { SidebarProvider, SidebarInset, SidebarTrigger } from '~/components/ui/sidebar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import { Separator } from '~/components/ui/separator'
import AppSidebar from '~/components/layout/AppSidebar.vue'
import AppBreadcrumbs from '~/components/layout/AppBreadcrumbs.vue'
import ThemeToggle from '~/components/layout/ThemeToggle.vue'

const { locale, setLocale, t } = useI18n()
const authStore = useAuthStore()
const orgStore = useOrgStore()

onMounted(async () => {
  if (orgStore.organizations.length === 0) {
    try {
      await orgStore.fetchOrganizations()
    }
    catch {
      // Non-critical for layout
    }
  }
})
</script>

<template>
  <!-- Desktop-only guard (D-07): screens below 1024px -->
  <div class="block min-[1024px]:hidden min-h-screen bg-background">
    <div class="flex min-h-screen flex-col items-center justify-center p-6 text-center">
      <h1 class="text-2xl font-semibold text-foreground">
        Desktop Required
      </h1>
      <p class="mt-3 max-w-md text-sm text-muted-foreground">
        This application requires a screen width of at least 1024px. Please use a desktop browser.
      </p>
    </div>
  </div>

  <!-- Main layout: hidden below 1024px -->
  <div class="hidden min-[1024px]:block">
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <!-- Top bar: 48px height -->
        <header class="flex h-12 shrink-0 items-center border-b bg-background px-4">
          <div class="flex items-center gap-2">
            <SidebarTrigger class="-ml-1" />
            <Separator orientation="vertical" class="mr-2 h-4" />
            <AppBreadcrumbs />
          </div>

          <div class="ml-auto flex items-center gap-1">
            <ThemeToggle />

            <!-- User avatar dropdown -->
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" class="h-8 w-8">
                  <Avatar class="h-7 w-7">
                    <AvatarFallback class="text-xs">
                      {{ authStore.user?.name?.charAt(0)?.toUpperCase() ?? 'U' }}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-56">
                <div class="px-2 py-1.5">
                  <p class="text-sm font-medium">{{ authStore.user?.name ?? 'User' }}</p>
                  <p class="text-xs text-muted-foreground">{{ authStore.user?.email ?? '' }}</p>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>{{ t('common.language') }}</DropdownMenuLabel>
                <DropdownMenuRadioGroup :model-value="locale" @update:model-value="setLocale">
                  <DropdownMenuRadioItem value="ru">
                    {{ t('language.ru') }}
                  </DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="en">
                    {{ t('language.en') }}
                  </DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
                <DropdownMenuSeparator />
                <DropdownMenuItem class="cursor-pointer gap-2" @click="authStore.logout()">
                  <LogOut class="size-4" />
                  <span>{{ t('auth.logout') }}</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        <!-- Content area: 24px padding (lg spacing token) -->
        <div class="p-6">
          <slot />
        </div>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template>

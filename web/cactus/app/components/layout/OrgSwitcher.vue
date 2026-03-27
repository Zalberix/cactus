<script setup lang="ts">
import { ChevronsUpDown, Building2, Check } from 'lucide-vue-next'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'
import { SidebarMenuButton } from '~/components/ui/sidebar'
import { useSidebar } from '~/components/ui/sidebar/utils'

const { t } = useI18n()
const orgStore = useOrgStore()
const { state } = useSidebar()

async function handleSwitch(orgId: number) {
  orgStore.switchOrg(orgId)
  await navigateTo(`/org/${orgId}/workflows`)
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <SidebarMenuButton
        size="lg"
        class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
      >
        <div class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Building2 class="size-4" />
        </div>
        <div v-if="state === 'expanded'" class="flex flex-col gap-0.5 leading-none">
          <span class="font-semibold truncate">
            {{ orgStore.currentOrg?.name ?? t('org.switch') }}
          </span>
        </div>
        <ChevronsUpDown v-if="state === 'expanded'" class="ml-auto size-4" />
      </SidebarMenuButton>
    </DropdownMenuTrigger>
    <DropdownMenuContent
      class="w-[--reka-dropdown-menu-trigger-width] min-w-56"
      align="start"
    >
      <DropdownMenuLabel>{{ t('org.switch') }}</DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuItem
        v-for="org in orgStore.organizations"
        :key="org.id"
        class="cursor-pointer gap-2"
        @click="handleSwitch(org.id)"
      >
        <Building2 class="size-4" />
        <span>{{ org.name }}</span>
        <Check
          v-if="org.id === orgStore.currentOrgId"
          class="ml-auto size-4"
        />
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

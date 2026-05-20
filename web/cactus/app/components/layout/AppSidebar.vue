<script setup lang="ts">
import {
  Workflow,
  Server,
  MessageSquare,
  Settings,
  Users,
  Shield,
  Box,
  KeyRound,
  ChevronDown,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-vue-next'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarSeparator,
} from '~/components/ui/sidebar'
import { useSidebar } from '~/components/ui/sidebar/utils'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '~/components/ui/collapsible'
import OrgSwitcher from './OrgSwitcher.vue'

const { t } = useI18n()
const route = useRoute()
const orgStore = useOrgStore()
const { state, toggleSidebar } = useSidebar()

const orgBase = computed(() =>
  orgStore.currentOrgId ? `/org/${orgStore.currentOrgId}` : '/org/0',
)

const navItems = computed(() => [
  {
    label: t('nav.workflows'),
    icon: Workflow,
    to: `${orgBase.value}/workflows`,
  },
  {
    label: t('nav.workers'),
    icon: Server,
    to: `${orgBase.value}/workers`,
  },
  {
    label: t('nav.messages'),
    icon: MessageSquare,
    to: `${orgBase.value}/messages`,
  },
])

const settingsItems = computed(() => [
  {
    label: t('nav.users'),
    icon: Users,
    to: `${orgBase.value}/settings/users`,
  },
  {
    label: t('nav.roles'),
    icon: Shield,
    to: `${orgBase.value}/settings/roles`,
  },
  {
    label: t('nav.systems'),
    icon: Box,
    to: `${orgBase.value}/settings/systems`,
  },
  {
    label: t('nav.bootstrapTokens'),
    icon: KeyRound,
    to: `${orgBase.value}/settings/worker-tokens`,
  },
])

function isActive(to: string) {
  return route.path.startsWith(to)
}

const isSettingsOpen = computed(() =>
  route.path.includes('/settings'),
)
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <OrgSwitcher />
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>

    <SidebarSeparator />

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>{{ t('nav.workflows') }}</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in navItems" :key="item.to">
              <SidebarMenuButton
                as-child
                :data-active="isActive(item.to)"
                :tooltip="item.label"
              >
                <NuxtLink :to="item.to">
                  <component :is="item.icon" />
                  <span>{{ item.label }}</span>
                </NuxtLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>

      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <Collapsible :default-open="isSettingsOpen" class="group/collapsible">
              <SidebarMenuItem>
                <CollapsibleTrigger as-child>
                  <SidebarMenuButton :tooltip="t('nav.settings')">
                    <Settings />
                    <span>{{ t('nav.settings') }}</span>
                    <ChevronDown
                      v-if="state === 'expanded'"
                      class="ml-auto size-4 transition-transform group-data-[state=open]/collapsible:rotate-180"
                    />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    <SidebarMenuSubItem v-for="item in settingsItems" :key="item.to">
                      <SidebarMenuSubButton
                        as-child
                        :data-active="isActive(item.to)"
                      >
                        <NuxtLink :to="item.to">
                          <component :is="item.icon" class="size-4" />
                          <span>{{ item.label }}</span>
                        </NuxtLink>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            :tooltip="state === 'expanded' ? 'Collapse' : 'Expand'"
            @click="toggleSidebar"
          >
            <ChevronsLeft v-if="state === 'expanded'" />
            <ChevronsRight v-else />
            <span v-if="state === 'expanded'">Collapse</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
  </Sidebar>
</template>

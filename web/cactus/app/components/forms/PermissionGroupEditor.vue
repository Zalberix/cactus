<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronDown } from 'lucide-vue-next'
import type { Permission } from '~/utils/permissions'
import { PERMISSION_GROUPS } from '~/utils/permissions'
import { Checkbox } from '~/components/ui/checkbox'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '~/components/ui/collapsible'
import { cn } from '~/utils/cn'

const props = defineProps<{
  modelValue: Permission[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Permission[]]
}>()

const { t } = useI18n()

const search = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const debouncedSearch = ref('')

function onSearchInput(value: string | number) {
  search.value = String(value)
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    debouncedSearch.value = search.value.toLowerCase()
  }, 150)
}

const filteredGroups = computed(() => {
  if (!debouncedSearch.value) return PERMISSION_GROUPS

  return PERMISSION_GROUPS
    .map(group => ({
      ...group,
      permissions: group.permissions.filter(
        p => p.label.toLowerCase().includes(debouncedSearch.value)
          || p.value.toLowerCase().includes(debouncedSearch.value),
      ),
    }))
    .filter(group => group.permissions.length > 0)
})

function isChecked(permission: Permission): boolean {
  return props.modelValue.includes(permission)
}

function togglePermission(permission: Permission) {
  const current = [...props.modelValue]
  const idx = current.indexOf(permission)
  if (idx >= 0) {
    current.splice(idx, 1)
  }
  else {
    current.push(permission)
  }
  emit('update:modelValue', current)
}

function groupSelectedCount(group: typeof PERMISSION_GROUPS[number]): number {
  return group.permissions.filter(p => props.modelValue.includes(p.value)).length
}

function groupCheckState(group: typeof PERMISSION_GROUPS[number]): boolean | 'indeterminate' {
  const selected = groupSelectedCount(group)
  if (selected === 0) return false
  if (selected === group.permissions.length) return true
  return 'indeterminate'
}

function toggleGroup(group: typeof PERMISSION_GROUPS[number]) {
  const current = [...props.modelValue]
  const state = groupCheckState(group)

  if (state === true) {
    // Deselect all in group
    const groupPerms = new Set(group.permissions.map(p => p.value))
    emit('update:modelValue', current.filter(p => !groupPerms.has(p)))
  }
  else {
    // Select all in group
    const groupPerms = group.permissions.map(p => p.value)
    const newValue = new Set([...current, ...groupPerms])
    emit('update:modelValue', [...newValue] as Permission[])
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Search -->
    <Input
      :model-value="search"
      :placeholder="t('roles.searchPermissions')"
      class="max-w-sm"
      @update:model-value="onSearchInput"
    />

    <!-- Permission Groups -->
    <div v-if="filteredGroups.length === 0" class="py-8 text-center text-sm text-muted-foreground">
      {{ t('common.noResults') }}
    </div>

    <div class="space-y-2">
      <Collapsible
        v-for="group in filteredGroups"
        :key="group.domain"
        :default-open="true"
        class="rounded-md border"
      >
        <template #default="{ open }">
          <!-- Group Header -->
          <CollapsibleTrigger as-child>
            <div
              class="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-muted/50"
            >
              <div class="flex items-center gap-3">
                <Checkbox
                  :model-value="groupCheckState(group)"
                  @click.stop
                  @update:model-value="toggleGroup(group)"
                />
                <span class="font-medium text-sm">{{ group.label }}</span>
                <span class="text-xs text-muted-foreground">
                  {{ t('roles.selectedCount', {
                    n: groupSelectedCount(group),
                    total: group.permissions.length,
                  }) }}
                </span>
              </div>
              <ChevronDown
                :class="cn(
                  'h-4 w-4 text-muted-foreground transition-transform duration-200',
                  open && 'rotate-180',
                )"
              />
            </div>
          </CollapsibleTrigger>

          <!-- Group Content -->
          <CollapsibleContent>
            <div class="border-t px-4 py-3 space-y-3">
              <div
                v-for="perm in group.permissions"
                :key="perm.value"
                class="flex items-center gap-3"
              >
                <Checkbox
                  :id="`perm-${perm.value}`"
                  :model-value="isChecked(perm.value)"
                  @update:model-value="togglePermission(perm.value)"
                />
                <Label :for="`perm-${perm.value}`" class="text-sm cursor-pointer">
                  {{ perm.label }}
                </Label>
              </div>
            </div>
          </CollapsibleContent>
        </template>
      </Collapsible>
    </div>
  </div>
</template>

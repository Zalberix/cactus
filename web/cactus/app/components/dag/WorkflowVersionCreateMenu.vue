<script setup lang="ts">
import { Copy, FilePlus2 } from 'lucide-vue-next'
import type { Version, VersionSummary } from '~/composables/useVersions'
import { Button } from '~/components/ui/button'
import { workflowVersionEditorPath } from '~/composables/useWorkflows'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '~/components/ui/dropdown-menu'

const props = defineProps<{
  workflowId: number
  orgId: number
  versions: VersionSummary[]
}>()

const emit = defineEmits<{
  created: [version: Version]
}>()

const router = useRouter()
const { createVersion, copyVersion } = useVersions()

const latestVersion = computed(() => props.versions[0])
const { t } = useI18n()

async function openCreated(version: Version) {
  emit('created', version)
  await router.push(workflowVersionEditorPath(props.orgId, props.workflowId, version.id))
}

async function createBlank() {
  await openCreated(await createVersion(props.workflowId))
}

async function copyLatest() {
  if (!latestVersion.value) return
  await openCreated(await copyVersion(latestVersion.value.id))
}
</script>

<template>
    <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button>{{ t('workflowVersions.createMenu.create') }}</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem @click="createBlank">
        <FilePlus2 class="mr-2 h-4 w-4" />
        {{ t('workflowVersions.createMenu.createBlank') }}
      </DropdownMenuItem>
      <DropdownMenuItem :disabled="!latestVersion" @click="copyLatest">
        <Copy class="mr-2 h-4 w-4" />
        {{ t('workflowVersions.createMenu.copyLatest') }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

<script setup lang="ts">
import { Copy, FilePlus2 } from 'lucide-vue-next'
import type { Version, VersionSummary } from '~/composables/useVersions'
import { Button } from '~/components/ui/button'
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

const latestVersion = computed(() => props.versions.find(v => !v.deleted_at))

async function openCreated(version: Version) {
  emit('created', version)
  await router.push(`/org/${props.orgId}/workflows/${props.workflowId}/edit?versionId=${version.id}`)
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
      <Button>Create Version</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem @click="createBlank">
        <FilePlus2 class="mr-2 h-4 w-4" />
        Create blank version
      </DropdownMenuItem>
      <DropdownMenuItem :disabled="!latestVersion" @click="copyLatest">
        <Copy class="mr-2 h-4 w-4" />
        Copy latest version
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

<template>
    <div id="row">
        <TokensEdit v-if="isVisibleEdit" />
        <TokensDelete v-if="isVisibleDelete" />
        <div class="selected" @click="toggleCheckboxes">
            <CheckboxStateOne v-if="showFirstCheckbox" />
            <CheckboxStateTwo v-else />
        </div>
        <div class="name" @click="toggleCheckboxes">
            <span>{{ name }}</span>
        </div>
        <div class="public-key">
            <span>{{ publicKey }}</span>
        </div>
        <div class="status">
            <span>{{ status }}</span>
        </div>
        <div class="description">
            <span>{{ description }}</span>
        </div>
        <div class="process">
            <span>{{ process }}</span>
        </div>
        <div class="action">
            <Trash @click="isVisTrash" />
            <EditPencil @click="isVisEdit" />
        </div>
    </div>
</template>

<script setup lang="ts">
import Trash from '@/components/icons/Trash.vue'
import EditPencil from '@/components/icons/EditPencil.vue'
import CheckboxStateOne from '@/components/ui/CheckboxStateOne.vue'
import CheckboxStateTwo from '@/components/ui/CheckboxStateTwo.vue'
import TokensEdit from '@/views/core/tokens/TokensEdit.vue'
import TokensDelete from '@/views/core/tokens/TokensDelete.vue'
import { ref } from 'vue'

const props = defineProps({
    name: String,
    publicKey: String,
    status: String,
    description: String,
    process: String
})

const showFirstCheckbox = ref(true)
const isVisibleEdit = ref(false)
const isVisibleDelete = ref(false)

const toggleCheckboxes = function () {
    showFirstCheckbox.value = !showFirstCheckbox.value
}

const isVisEdit = function () {
    isVisibleEdit.value = !isVisibleEdit.value
}
const isVisTrash = function () {
    isVisibleDelete.value = !isVisibleDelete.value
}
</script>

<style scoped>
@import '../../../src/assets/css/TokensRow.css';
</style>

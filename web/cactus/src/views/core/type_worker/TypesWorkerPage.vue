<template>
    <div>
        <div class="main_content">
            <HeaderPages :info_header="breadcrumbs" />
            <div class="content_block">
                <div v-for="element of blockName" :key="element.id" class="block">
                    <RouterLink :to="{ name: element.slug }">
                        <span class="text">{{ element.name }}</span>
                    </RouterLink>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import HeaderPages from '@/components/HeaderPages.vue'
import { ref, onMounted } from 'vue'
import axios from 'axios'

interface InfoHeader {
    url_info: string
    url: string
    name_pages?: string
}

const breadcrumbs: InfoHeader[] = [
    { url_info: 'Главная', url: '/', name_pages: 'Процессы' },
    { url_info: 'Типы работников', url: '/processes' },
    { url_info: 'Рассылка писем', url: '/emails' }
]

interface Process {
    id: number
    name: string
    slug: string
}

const blockName = ref<Process[]>([])

const fetchProcesses = async () => {
    try {
        const response = await axios.post('http://localhost:8080/api/app/types-worker/list/')
        blockName.value = response?.data?.data?.types_worker ?? []
    } catch (error) {
        console.error('Error fetching processes:', error)
    }
}

onMounted(() => {
    fetchProcesses()
})
</script>

<style>
.main_content {
    width: 100%;
}
.content_block {
    width: 100%;
    margin: 10px 0;
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
}

.block {
    width: 300px;
    height: 80px;
    padding: 10px;
    border: 1px solid rgba(146, 147, 171, 0.5);
    border-radius: 10px;
    display: flex;
    justify-content: center;
    align-items: center;
    box-sizing: border-box;
}

.text {
    font-family: Inter;
    font-weight: 400;
    font-size: 20px;
    line-height: 29px;
    color: black;
}
</style>

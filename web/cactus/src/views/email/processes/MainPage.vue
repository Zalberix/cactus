<template>
    <div>
        <div class="main_content">
            <Header_Pages :info_header="info_header" />
        </div>
    </div>
</template>

<script setup>
import Header_Pages from '../components/Header_Pages.vue'

const info_header = [
    { url_info: 'Cactus', url: '/cactus', name_pages: 'Процессы' },
    { url_info: 'Процессы', url: '/processes' },
    { url_info: 'Рассылка писем', url: '/emails' }
]
</script>

<style></style>

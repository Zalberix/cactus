<template>
    <div class="body">
        <NavbarPanel class="navigation" />
        <RouterView class="content" />
    </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavbarPanel from '@/components/NavbarPanel.vue'
</script>

<style scoped lang="less">
.body {
    display: flex;
}


// * {
//     margin: 0;
//     padding: 0;
//     box-sizing: border-box;
//     font-size: 14px;
//     font-family: 'Inter', sans-serif;
// }

// @font-face {
//     font-family: 'Inter';
//     src: url('../font/Inter-VariableFont_slnt,wght.ttf');
// }

// a {
//     color: #fff;
//     text-decoration: none;
// }

// a:hover {
//     color: #fff;
//     text-decoration: none;
// }

// a:visited {
//     color: #fff;
//     text-decoration: none;
// }

// .app {
//     display: flex;
// }

// .main_content {
//     width: max-content;
//     height: max-content;
//     margin-left: 199px;
// }
</style>

<template>
    <div class="main_container">
        <div class="content">
            <div class="header">
                <p>Процессы</p>
                <h2>Рассылка писем</h2>
            </div>
            <div class="main">
                <RouterLink
                    v-for="email in emails"
                    :key="email.id"
                    class="email_info_clik"
                    :to="'/process/email/list'"
                >
                    <div class="email_info">
                        <div class="icon" v-html="email.icon"></div>
                        <div class="client_name style_and_position">
                            <p>{{ email.clientName }}</p>
                        </div>
                        <div class="style_and_position">
                            <p>{{ email.subject }}</p>
                        </div>
                        <div class="right_info">
                            <div class="style_and_position">
                                <p id="date">{{ email.date }}</p>
                            </div>
                            <div class="style_and_position">
                                <DotIcon />
                            </div>
                            <div class="time style_and_position">
                                <p>{{ email.time }}</p>
                            </div>
                        </div>
                    </div>
                </RouterLink>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import DotIcon from '@/components/icons/DotIcon.vue'
import Wait from '@/components/icons/WaitStatusPipline.vue'
import Execution from '@/components/icons/ExecutionStatusPipline.vue'
import Success from '@/components/icons/SuccessStatusPipline.vue'
import Cancel from '@/components/icons/CancelStatusPipline.vue'
import Error from '@/components/icons/ErrorStatusPipline.vue'
import { ref, type Component } from 'vue'

interface EmailInfo {
    id: number
    icon: Component
    clientName: string
    subject: string
    date: string
    time: string
}

const emails = ref<EmailInfo[]>([
    {
        id: 1,
        icon: Wait,
        clientName: 'info.agt72.ru',
        subject: 'Новости за 07.06.2024',
        date: 'Отправка 07.06.2024 09:00:00',
        time: '00:31'
    },
    {
        id: 2,
        icon: Execution,
        clientName: 'agt72.ru',
        subject: 'Новости за 02.06.2024',
        date: 'только что',
        time: '00:01'
    },
    {
        id: 3,
        icon: Success,
        clientName: 'dom.tyumen-city',
        subject: 'Востановление пароля',
        date: '5 минут назад',
        time: '12:01'
    },
    {
        id: 4,
        icon: Error,
        clientName: 'uslugi.tymen-city.ru',
        subject: 'Отправка заявок по ярмаркам',
        date: '15 минут назад',
        time: '07:17'
    },
    {
        id: 5,
        icon: Cancel,
        clientName: 'children.tyumen-city.ru',
        subject: 'Рассылка дипломов по конкурсу “Тюменская каравелла”',
        date: '16 часов назад',
        time: '00:31'
    },
    {
        id: 6,
        icon: Success,
        clientName: 'leto.tyumen-city.ru',
        subject: 'Новости за 01.06.2024',
        date: '5 минут назад',
        time: '12:01'
    }
])
</script>

<style scoped lang="less">
* {
    padding: 0;
    margin: 0;
    box-sizing: border-box;
}

@font-face {
    font-family: 'Inter';
    src: url('../font/Inter-VariableFont_slnt,wght.ttf');
}

.content {
    width: 84%;
    height: 100vh;
    margin-left: 239px;
    position: absolute;
    background-color: rgba(250, 251, 252, 1);
}
.header {
    padding: 15px;
}
.header p {
    color: rgba(146, 147, 171, 1);
    font-size: 12px;
    font-family: 'Inter';
}
.header p:hover {
    text-decoration: underline;
    cursor: pointer;
}
.header h2 {
    font-family: 'Inter';
    font-size: 32px;
}
.email_info {
    width: 95%;
    display: flex;
    position: relative;
    padding: 2px 8px 0 8px;
    margin: 15px;
    background-color: rgba(255, 255, 255, 1);
    border-radius: 10px;
    border: 1px solid rgba(146, 147, 171, 0.5);
}
.icon {
    margin: 3px 0;
}
.client_name {
    margin: 5px 15px;
    height: 20px;
    padding: 1px 5px 0 5px;
    border-radius: 10px;
    background-color: rgba(146, 147, 171, 1);
}
.client_name p {
    color: #fff;
    font-family: 'Inter';
    font-size: 15px;
}
.style_and_position {
    margin: 7px 15px;
    font-family: 'Inter';
}
.right_info {
    right: 0;
    position: absolute;
    display: flex;
}
.time {
    width: 45px;
}
.email_info_clik {
    text-decoration: none;
    color: black;
}
.email_info:hover {
    background-color: rgba(230, 230, 234, 0.5);
    transition: 0.3s;
}
</style>

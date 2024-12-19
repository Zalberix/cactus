<template>
    <div>
        <div class="main_container">
            <div class="content">
                <div class="header">
                    <p>Процессы</p>
                    <h2>Рассылка писем</h2>
                </div>
                <div class="main">
                    <RouterLink
                        v-for="(email, index) in emailData"
                        :key="index"
                        class="email_info_clik"
                        :to="email.link"
                    >
                        <div class="email_info">
                            <div class="icon">
                                <component :is="email.icon" />
                            </div>
                            <div class="client_name style_and_position">
                                <p>{{ email.clientName }}</p>
                            </div>
                            <div class="style_and_position">
                                <p>{{ email.newsDate }}</p>
                            </div>
                            <div class="right_info">
                                <div class="style_and_position">
                                    <p id="date">{{ email.sentDate }}</p>
                                </div>
                                <div class="style_and_position">
                                    <DotIcon class="icon" />
                                </div>
                                <div class="time style_and_position">
                                    <p>{{ email.duration }}</p>
                                </div>
                            </div>
                        </div>
                    </RouterLink>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import DotIcon from '@/components/icons/DotIcon.vue'
import Wait from '@/components/icons/WaitStatusPipline.vue'
import Execution from '@/components/icons/ExecutionStatusPipline.vue'
import Success from '@/components/icons/SuccessStatusPipline.vue'
import Cancel from '@/components/icons/CancelStatusPipline.vue'
import Error from '@/components/icons/ErrorStatusPipline.vue'

const emailData = [
    {
        link: '/process/email/list',
        icon: Wait,
        clientName: 'info.agt72.ru',
        newsDate: 'Новости за 07.06.2024',
        sentDate: 'Отправка 07.06.2024 09:00:00',
        duration: '00:31'
    },
    {
        link: '/process/email/list',
        icon: Execution,
        clientName: 'agt72.ru',
        newsDate: 'Новости за 02.06.2024',
        sentDate: 'только что',
        duration: '00:01'
    },
    {
        link: '/process/email/list',
        icon: Success,
        clientName: 'dom.tyumen-city',
        newsDate: 'Востановление пароля',
        sentDate: '5 минут назад',
        duration: '12:01'
    },
    {
        link: '/process/email/list',
        icon: Error,
        clientName: 'uslugi.tymen-city.ru',
        newsDate: 'Отправка заявок по ярмаркам',
        sentDate: '15 минут назад',
        duration: '07:17'
    },
    {
        link: '/process/email/list',
        icon: Cancel,
        clientName: 'children.tyumen-city.ru',
        newsDate: 'Рассылка дипломов по конкурсу "Тюменская каравелла"',
        sentDate: '16 часов назад',
        duration: '00:31'
    },
    {
        link: '/process/email/list',
        icon: Success,
        clientName: 'leto.tyumen-city.ru',
        newsDate: 'Новости за 01.06.2024',
        sentDate: '5 минут назад',
        duration: '12:01'
    }
]
</script>

<style scoped>
* {
    padding: 0;
    margin: 0;
    box-sizing: border-box;
}

@font-face {
    font-family: 'Inter';
    src: url('../font/Inter-VariableFont_slnt\,wght.ttf');
}

.content {
    width: 84%;
    padding: 20px;
    margin-left: 239px;
    position: absolute;
}

.process {
    display: flex;
}

#link {
    text-decoration: none;
    color: black;
}

#link:hover {
    text-decoration: underline;
    transition: 3s;
}

.process svg {
    margin-top: 2px;
    margin-right: 10px;
}

.process p {
    color: rgba(146, 147, 171, 1);
    margin-right: 10px;
    font-size: 15px;
}
.process p:hover {
    text-decoration: underline;
    cursor: pointer;
}

.name_client {
    display: flex;
    margin-top: 20px;
}

.name_client svg {
    margin-top: 6px;
    margin-right: 15px;
}

.name_client h2 {
    font-size: 32px;
    font-family: 'Inter';
}

.main_information_container {
    display: flex;
    margin-top: 30px;
}
.main_information h2 {
    font-family: 'Inter';
    margin-top: 20px;
    margin-left: 5px;
    font-size: 32px;
}
#name_block {
    display: in;
}

.info {
    display: flex;
}

.style {
    margin: 5px;
    width: 180px;
    height: 70px;
    padding: 15px;
    background-color: #fff;
    border: 1px solid rgba(146, 147, 171, 0.5);
    border-radius: 10px;
}

.style p {
    text-align: center;
    color: rgba(146, 147, 171, 1);
    font-size: 12px;
    font-family: 'Inter';
    margin-bottom: 10px;
}

.style h3 {
    text-align: center;
    font-family: 'Inter';
    font-size: 13px;
    font-weight: 600;
    margin: 0 auto;
}

#color_text {
    color: green;
}

.chart_diogram {
    margin-top: 5px;
    display: flex;
    align-items: flex-end;
    height: 150px;
    padding: 7px;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    background: #fff;
}

.num_bar p {
    color: rgba(146, 147, 171, 1);
    text-align: center;
    font-size: 10px;
    font-family: 'Intet';
    margin-bottom: 12px;
    margin-top: 12px;
}

.main_bar {
    width: 5px;
    height: 120px;
    border-radius: 10px;
    background-color: rgba(146, 147, 171, 0.25);
    margin: 5px;
    position: relative;
}

.bar {
    width: 5px;
    height: 70px;
    position: absolute;
    border-radius: 10px;
    bottom: 0;
}

.resout_1 {
    width: 5px;
    height: 10px;
    border-radius: 10px;
}

.bar.green {
    background-color: rgba(39, 189, 104, 1);
}
.bar.orange {
    background-color: rgba(246, 172, 85, 1);
}
.bar.red {
    background-color: rgba(246, 85, 95, 1);
}
.bar.blue {
    background-color: rgba(85, 111, 246, 1);
}
.bar.gray {
    background-color: rgba(146, 147, 171, 0.25);
}
.main_content h2 {
    font-family: 'Inter';
    font-size: 32px;
    margin-left: 5px;
    margin-top: 15px;
}
.email_info {
    width: 99%;
    display: flex;
    position: relative;
    padding: 2px 8px 0 8px;
    margin: 20px 0 0 5px;
    background-color: rgba(255, 255, 255, 1);
    border-radius: 10px;
    border: 1px solid rgba(146, 147, 171, 0.5);
}

.icon {
    margin: 3px 0;
}
.number {
    font-family: 'Inter';
    margin-top: 7px;
    color: rgba(146, 147, 171, 1);
}

.client_name {
    display: flex;
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

#uuid {
    margin-left: 7px;
}

.style_and_position {
    margin: 7px 15px;
    font-family: 'Inter';
}
.name_and_date p {
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

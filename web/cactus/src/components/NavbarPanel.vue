<template>
    <div class="navigation">
        <div class="header">
            <div class="logo_and_name">
                <div class="logo">
                    <LogoIcon />
                </div>
                <div class="name">
                    <RouterLink to="/">Cactus</RouterLink>
                </div>
            </div>
        </div>

        <div class="container">
            <NavbarButton to="/" :icon="DashboardIcon" title="Панель управления" />
            <NavbarButton to="/processes/" :icon="ProcessIcon" title="Процессы" v-slot="{to}">
                <RouterLink
                    v-for="subTo of listSubProcessButton"
                    :key="subTo"
                    :class="{ active: activeRouter(`${to}${subTo}/`) }"
                    class="subbutton"
                    :to="`${to}${subTo}/`"
                >
                    <DotIcon class="icon" />
                    <span>{{ subTo }}</span>
                </RouterLink>
            </NavbarButton>
            <NavbarButton to="/users/" :icon="UsersIcon" title="Пользователи" />
            <NavbarButton to="/tokens/" :icon="TokenIcon" title="Токены" />
            <NavbarButton to="/tools/" :icon="ToolsIcon" title="Настройки" />
            <NavbarButton to="/docs/" :icon="DocumentationIcon" title="Документация" />
        </div>

        <div class="footer">
            <div class="logo_prof_and_name_user_and_exit">
                <div class="logo_user">
                    <img src="../src/assets/img/ava.png" alt="" />
                </div>
                <div class="name_user">
                    <RouterLink class="user_name_text" to="/">Zalberix</RouterLink>
                </div>
                <div class="exit">
                    <RouterLink class="user_name_text" to="/">
                        <ExitIcon />
                    </RouterLink>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import LogoIcon from './icons/LogoIcon.vue'
import ExitIcon from './icons/ExitIcon.vue'
import DashboardIcon from './icons/DashboardIcon.vue'
import ProcessIcon from './icons/ProcessIcon.vue'
import DocumentationIcon from './icons/DocumentationIcon.vue'
import ToolsIcon from './icons/ToolsIcon.vue'
import TokenIcon from './icons/TokenIcon.vue'
import UsersIcon from './icons/UsersIcon.vue'
import DotIcon from './icons/DotIcon.vue'
import NavbarButton from './ui/NavbarButton.vue'

import { useRoute } from 'vue-router'

const route = useRoute()

const listSubProcessButton = ['email', 'telegram', 'sms', 'push', '1c']

function activeRouter(url: string) {
    return route.path === url
}
</script>

<style lang="less" scope>
.navigation {
    min-width: 210px;
    max-width: 12vw;
    height: 100vh;
    background-color: rgba(17, 20, 61, 1);
    display: flex;
    flex-direction: column;

    .header {
        height: 70px;
        background-color: rgba(41, 43, 80, 1);
        display: flex;
        align-items: center;

        .logo_and_name {
            height: 100%;
            padding: 22px;
            display: flex;
            position: relative;
            align-items: center;
            gap:10px;
            .logo {
                width: 40px;
                height: 40px;
                border-radius: 50%;
                background-color: #fff;
                svg {
                    margin: 1px 0 0 8px;
                }
            }
            .name {
                color: #fff;
                text-align: center;
                font-family: 'Inter';
                text-transform: uppercase;
            }
        }


    }

    .container {
        flex-grow: 1;
        overflow: auto;
        padding: 15px;
        max-width: 100%;
    }

    .footer {
        width: 100%;
        height: 70px;
        display: flex;
        justify-content: center;
        align-items: center;
        background-color: rgba(41, 43, 80, 1);

        .logo_prof_and_name_user_and_exit {
            // width: 200px;
            // margin: auto;
            display: flex;
            .logo_user {
                padding: 10px;
                img {
                    width: 40px;
                    height: 40px;
                    border-radius: 50%;
                }
            }
            .name_user {
                margin-top: 20px;
                margin-left: 20px;
                color: #fff;
                .user_name_text {
                    color: #fff;
                    text-decoration: none;
                }
            }
            .exit {
                margin-top: 20px;
                margin-left: 20px;
                .user_name_text {
                    color: #fff;
                    text-decoration: none;
                }
            }
        }
    }   
}
</style>

<template>
    <div class="navbar_buttons" :class="{ open: open }">
        <div :class="{ active: isExactActive(to) }" class="header_button">
            <RouterLink class="link" :to="to">
                <div class="content"><icon />{{ title }}</div>
            </RouterLink>
            <div v-if="$slots.default" @click="open = !open" class="roll_button">
                <ArrowIcon class="roll_icon" :class="{ open: open }" />
            </div>
        </div>
        <div v-if="$slots.default && open" class="subbutton_list">
            <slot :to="to"></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import ArrowIcon from '../icons/ArrowIconR.vue'

import { ref, type Component } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

interface Props {
    to: string
    icon: Component
    title: string
}

defineProps<Props>()
const open = ref(false)

function isExactActive(url: string) {
    return (
        route.path === url ||
        (url.split('/').filter((e) => e !== '').length !== 0 && route.path.startsWith(url))
    )
}
</script>

<style lang="less" scope>
.navbar_buttons {
    width: 100%;
    color: #fff;
    font-size: 12px;
    font-family: 'Inter';
    margin: 10px 0;
    border-radius: 8px;
    cursor: pointer;
    &:first-child {
        margin-top: 0;
        margin-bottom: 10px;
    }
    &:last-child {
        margin-top: 10px;
        margin-bottom: 0;
    }
    .header_button {
        padding: 7px 10px;
        border-radius: 8px;
        cursor: pointer;

        display: flex;
        justify-content: space-between;
        text-align: center;
        width: 100%;

        &:hover {
            background-color: rgba(85, 111, 246, 1);
        }

        &.active {
            background-color: rgba(85, 111, 246, 1);
        }
        .link {
            width: 100%;
            flex-grow: 1;
        }
        .content {
            display: flex;
            align-items: center;
            gap: 10px;

            .icon {
                display: flex;
                justify-content: center;
                align-items: center;
                width: 20px;
                height: 20px;
            }
        }
    }
    &.open {
        background-color: rgba(85, 111, 246, 0.1);
    }

    .roll_button {
        width: 10px;
        display: flex;
        align-items: center;
        .roll_icon {
            color: #fff;
            transition: 0.2s;
            transform: rotate(90deg);
            display: flex;
            justify-items: center;
            &.open {
                transform: rotate(-90deg);
            }
        }
    }

    .subbutton_list {
        display: flex;
        flex-direction: column;
        gap: 10px;
        border-radius: 10px;
        transition: 0.5s;
        padding: 10px;

        .subbutton {
            padding: 3px 8px;
            color: #fff;
            display: flex;
            align-items: center;
            gap: 16px;

            &:hover,
            &:focus-visible {
                color: #fff;
                .icon {
                    color: #fff;
                }
            }

            &.active {
                color: rgba(85, 111, 246, 1);
                border-radius: 5px;
                .icon {
                    color: rgba(85, 111, 246, 1);
                }
            }

            .icon {
                color: #fff;
            }
        }
    }
}
</style>

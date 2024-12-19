<template>
    <div class="backgroundCol">
        <div class="main-block">
            <div class="header">
                <div class="edit-pencil">
                    <Trash style="cursor: default" />
                    <span id="name-edit">Удаление</span>
                </div>
                <div class="cross">
                    <Cross />
                </div>
            </div>
            <div class="main-content">
                <div>
                    <p>Вы действительно хотите удалить токен agt72.ru</p>
                </div>
                <div id="buttons">
                    <button id="eReset">
                        <span class="buttons-text">Отмена</span>
                    </button>
                    <button id="createToken">
                        <span class="buttons-text">Изменить</span>
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import Trash from '@/components/icons/Trash.vue'
import Cross from '@/components/icons/CrossIcon.vue'
</script>

<style scoped lang="less">
.backgroundCol {
    position: fixed;
    top: 0;
    left: 0;
    height: 100%;
    width: 100%;
    background-color: rgba(22, 44, 67, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1;

    .main-block {
        background-color: white;
        width: 384px;
        height: 148px;
        border-radius: 8px;
        padding: 15px;
        display: flex;
        flex-direction: column;
        gap: 10px;

        .header {
            width: 100%;
            height: 24px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            color: rgba(22, 44, 67, 1);

            .edit-pencil {
                display: flex;
                align-items: center;
                gap: 10px;

                #name-edit {
                    font-size: 20px;
                }
            }

            .cross {
                // стили для крестика, если нужно добавить
            }
        }

        .main-content {
            width: 100%;
            height: 100%;
            color: #162c43;
            display: flex;
            flex-direction: column;
            gap: 20px;

            #buttons {
                width: 100%;
                display: flex;
                gap: 10px;
                justify-content: center;
                align-items: center;

                #eReset {
                    width: 100%;
                    height: 37px;
                    border-radius: 8px;
                    border: 1px solid rgba(246, 85, 95, 0.5);
                    color: rgba(246, 85, 95, 1);
                    background-color: white;
                    padding: 10px 20px;
                    cursor: pointer;
                }

                #createToken {
                    width: 100%;
                    height: 37px;
                    border-radius: 8px;
                    padding: 10px 20px;
                    background-color: rgba(85, 111, 246, 1);
                    border: none;
                    color: white;
                    cursor: pointer;
                }
            }

            .buttons-text {
                width: 68px;
                height: 17px;
                font-family: Inter;
                font-weight: 700;
                font-size: 14px;
            }
        }
    }
}
</style>

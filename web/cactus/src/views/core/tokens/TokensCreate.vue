<template>
    <div id="main-container">
        <header id="header">
            <div id="breadcrumb">
                <p id="cactus">cactus</p>
                <RightScobka />
                <RouterLink to="/tokens">
                    <p id="cactus">Токены</p>
                </RouterLink>
            </div>
            <div id="title">
                <RouterLink to="/tokens">
                    <div id="back-button"><BackTokensCreate /></div>
                </RouterLink>
                <p id="tokens">Создание токена</p>
            </div>
        </header>
        <main id="main-content">
            <form action="#" name="forma" id="forma">
                <div id="fields">
                    <div class="field">
                        <label for="" class="label-style">Публичный токен</label>
                        <input type="text" />
                    </div>
                    <div class="field">
                        <label for="" class="label-style">Приватный токен</label>
                        <input type="password" />
                    </div>
                    <div class="field">
                        <label for="" class="label-style">Название</label>
                        <input type="text" />
                    </div>
                    <div class="field">
                        <label for="" class="label-style">Описание</label>
                        <input type="text" />
                    </div>
                    <div class="process-field">
                        <div class="label-style label-noh">Процессы</div>
                        <div id="checkbox-container">
                            <div class="flex-checkbox" id="email">
                                <CheckboxStateOne />
                                <CheckboxStateTwo />
                                <span>email</span>
                            </div>
                            <div class="flex-checkbox" id="telegram">
                                <CheckboxStateOne />
                                <CheckboxStateTwo />
                                <span>telegram</span>
                            </div>
                            <div class="flex-checkbox" id="sms">
                                <CheckboxStateOne />
                                <CheckboxStateTwo />
                                <span>sms</span>
                            </div>
                            <div class="flex-checkbox" id="C1">
                                <CheckboxStateOne />
                                <CheckboxStateTwo />
                                <span>1C</span>
                            </div>
                            <div class="flex-checkbox" id="push">
                                <CheckboxStateOne />
                                <CheckboxStateTwo />
                                <span>push</span>
                            </div>
                        </div>
                    </div>
                </div>
                <div id="buttons">
                    <button id="eReset">
                        <span class="buttons-text">Очистить</span>
                    </button>
                    <button id="createToken">
                        <span class="buttons-text">Создать</span>
                    </button>
                </div>
            </form>
        </main>
    </div>
</template>

<script setup>
import BackTokensCreate from '@/components/icons/BackTokensCreate.vue'
import RightScobka from '@/components/icons/RightScobka.vue'
import CheckboxStateOne from '@/components/ui/CheckboxStateOne.vue'
import CheckboxStateTwo from '@/components/ui/CheckboxStateTwo.vue'
</script>

<style scoped lang="less">
#main-container {
    width: inherit;
    height: 982px;
    padding: 15px;
    margin-left: 199px;
    background-color: white;
    font-family: Inter;

    #header {
        height: 64px;
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 5px;

        #breadcrumb {
            width: inherit;
            height: 20px;
            display: flex;
            flex-direction: row;
            gap: 5px;

            #cactus {
                width: 40px;
                height: 15px;
                text-align: left;
                color: #9293ab;
                font-size: 12px;
                font-weight: 700;
                line-height: 14.52px;
            }
        }

        #title {
            width: inherit;
            height: 39px;
            display: flex;
            justify-content: flex-start;
            align-items: center;

            #back-button {
                width: max-content;
                height: max-content;
                padding: 0 5px;

                #back-button--svg {
                    width: 30px;
                    height: 30px;
                }
            }

            #tokens {
                width: max-content;
                height: inherit;
                font-weight: 700;
                font-size: 32px;
                line-height: 38.73px;
                color: #162c43;
            }
        }
    }

    #main-content {
        width: 100%;
        height: 873px;
        color: #162c43;

        #forma {
            display: flex;
            flex-direction: column;
            gap: 15px;

            #fields {
                width: inherit;
                height: 330px;
                padding: 10px;
                display: flex;
                flex-direction: column;
                gap: 10px;

                .field {
                    width: inherit;
                    height: 30px;
                    display: flex;
                    align-items: center;
                    gap: 10px;

                    input {
                        width: 100%;
                        height: 30px;
                        padding: 5px 10px;
                        border: 1px solid rgba(146, 147, 171, 0.5);
                        border-radius: 4px;
                    }
                }

                .process-field {
                    width: inherit;
                    height: 150px;
                    display: flex;
                    align-items: flex-start;
                    gap: 10px;

                    .label-style {
                        width: 150px;
                        height: 30px;
                        font-family: Inter;
                        font-weight: 400;
                        font-size: 14px;
                        line-height: 17px;
                        display: flex;
                        justify-content: flex-start;
                        align-items: center;

                        &.label-noh {
                            height: max-content;
                        }
                    }

                    #checkbox-container {
                        width: 100%;
                        height: inherit;
                        display: flex;
                        flex-direction: column;
                        flex-wrap: wrap;
                        row-gap: 5px;
                        column-gap: 25px;

                        .flex-checkbox {
                            display: flex;
                            align-items: center;
                            gap: 5px;
                        }
                    }
                }
            }

            #buttons {
                width: 1103px;
                margin-right: 10px;
                display: flex;
                gap: 10px;
                justify-content: flex-end;

                #eReset {
                    width: 108px;
                    height: 37px;
                    border-radius: 8px;
                    border: 1px solid rgba(146, 147, 171, 1);
                    color: rgba(146, 147, 171, 1);
                    background-color: white;
                    padding: 10px 20px;
                    cursor: pointer;
                }

                #createToken {
                    width: 108px;
                    height: 37px;
                    border-radius: 8px;
                    padding: 10px 20px;
                    background-color: rgba(85, 111, 246, 1);
                    border: none;
                    color: white;
                    cursor: pointer;
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
}
</style>

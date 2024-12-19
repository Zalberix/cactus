<template>
    <div class="backgroundCol">
        <div class="main-block">
            <div class="header">
                <div class="edit-pencil">
                    <EditPencil style="cursor: default" />
                    <span id="name-edit">Редактирование - agt72.ru</span>
                </div>
                <div class="cross">
                    <Cross />
                </div>
            </div>
            <div class="main-content">
                <form action="#" name="forma" id="forma">
                    <div id="fields">
                        <div class="field">
                            <label for="" class="label-style">Название</label>
                            <input type="text" />
                        </div>
                        <div class="field">
                            <label for="" class="label-style">Описание</label>
                            <input type="password" />
                        </div>
                        <div class="process-field" style="height: max-content">
                            <div class="label-style label-noh">Статус</div>
                            <div id="checkbox-container">
                                <div class="flex-checkbox" id="email">
                                    <CheckboxStateOne />
                                    <CheckboxStateTwo />
                                    <span>Активен</span>
                                </div>
                            </div>
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
                            <span class="buttons-text">Отмена</span>
                        </button>
                        <button id="createToken">
                            <span class="buttons-text">Изменить</span>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<style scoped>
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
}

.main-block {
    background-color: white;
    width: 605px;
    height: 590px;
    border-radius: 8px;
    padding: 15px;
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.header {
    width: 100%;
    height: 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: rgba(22, 44, 67, 1);
}

.edit-pencil {
    display: flex;
    align-items: center;
    gap: 10px;
}

#name-edit {
    font-size: 20px;
}

#main-content {
    width: 100%;
    height: 100%;
    color: #162c43;
}

#forma {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

#fields {
    width: inherit;
    height: 478px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 15px;
    color: rgba(22, 44, 67, 1);
}

.field {
    width: inherit;
    height: 30px;
    display: flex;
    align-items: center;
    gap: 10px;
}

.process-field {
    width: inherit;
    height: 150px;
    display: flex;
    align-items: flex-start;
    gap: 10px;
}

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
}

.label-noh {
    height: max-content;
}

.field input {
    width: 100%;
    height: 30px;
    padding: 5px 10px 5px 10px;
    border: 1px solid rgba(146, 147, 171, 0.5);
    border-radius: 4px;
}

#checkbox-container {
    width: 100%;
    height: inherit;
    display: flex;
    flex-direction: column;
    flex-wrap: wrap;
    row-gap: 5px;
    column-gap: 25px;
}

.flex-checkbox {
    display: flex;
    align-items: center;
    gap: 5px;
}

#buttons {
    width: 1103;
    margin-right: 10px;
    display: flex;
    gap: 10px;
    justify-content: flex-end;
}

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
</style>

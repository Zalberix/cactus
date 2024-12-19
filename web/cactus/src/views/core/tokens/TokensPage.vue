<script setup lang="ts">
import TokenRow from '@/views/core/tokens/TokensRow.vue';
import { RouterLink } from 'vue-router';

interface Props {
  isVisibleEdit: boolean;
  isVisibleDelete: boolean;
}

defineProps<Props>();
</script>

<template>
  <div id="main-container">
    <!--<TokensEdit />-->
    <header id="header">
      <div id="breadcrumb">
        <p id="cactus">cactus</p>
      </div>
      <div id="title">
        <p id="tokens">Токены</p>
        <RouterLink to="/tokens-create">
          <div id="create-token">
            <span id="create-token--text">Создать токен</span>
          </div>
        </RouterLink>
      </div>
    </header>
    <main id="main-content">
      <div id="content">
        <div id="table">
          <div id="table-header">
            <div class="selected">
              <span id="selected-text">Выбор</span>
            </div>
            <div class="name">
              <span id="name-text">Название</span>
            </div>
            <div class="public-key">
              <span id="public-key--text">Публичный токен</span>
            </div>
            <div class="status">
              <span id="status-text">Статус</span>
            </div>
            <div class="description">
              <span id="description-text">Описание</span>
            </div>
            <div class="process">
              <span id="process-text">Процессы</span>
            </div>
            <div class="action"></div>
          </div>
          <div id="header-line"></div>
          <div id="table-body">
            <TokenRow
              name="agt72.ru"
              public-key="jkbdmuyqqbojnerf"
              status="Активный"
              description="Главный сайт администрации"
              process="email, telegram"
            />
            <TokenRow />
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped lang="less">
#main-container {
  width: inherit;
  padding: 15px;
  background-color: white;
  margin-left: 199px;
}

#header {
  height: 64px;
  width: 100%;
  display: flex;
  flex-direction: column;
}

#breadcrumb {
  width: inherit;
  height: 20px;
}

#cactus {
  width: 40px;
  height: 15px;
  text-align: left;
  color: #9293ab;
  font-size: 12px;
  font-weight: 700;
  line-height: 14.52px;
}

#title {
  width: inherit;
  height: 39px;
  display: flex;
  justify-content: space-between;
}

#tokens {
  width: 121px;
  font-weight: 700;
  font-size: 32px;
  line-height: 38.73px;
  color: #162c43;
}

#create-token {
  width: 162px;
  height: 37px;
  border-radius: 8px;
  border: none;
  padding: 10px 20px;
  background-color: #556ff6;
  display: flex;
}

#create-token--text {
  width: 122px;
  height: 17px;
  font-weight: 700;
  font-size: 14px;
  line-height: 16.94px;
  color: white;
  text-transform: uppercase;
  text-decoration: none;
}

#main-content {
  width: 100%;
  height: 100%;
}

#content {
  width: inherit;
  height: 100%;
}

#table {
  height: 100%;
  border: 1px solid rgba(146, 147, 171, 0.5);
  border-radius: 10px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

#table-header {
  width: inherit;
  height: 20px;
  display: flex;
  gap: 10px;
}

#header-line {
  width: inherit;
  border: 1px solid rgba(146, 147, 171, 0.5);
}

#table-body {
  width: inherit;
  display: flex;
  flex-direction: column;
  gap: 15px;
}
</style>

<style scoped>
@import '../assets/css/TokensRow.css';
</style>
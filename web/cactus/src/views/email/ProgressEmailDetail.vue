<template>
    <div>
        <div class="main_container">
            <div class="content">
                <div class="status_bar">
                    <div class="done">
                        <div class="icon_and_time">
                            <div class="icon">
                                <svg
                                    width="30"
                                    height="30"
                                    viewBox="0 0 25 25"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <circle cx="12.5" cy="12.5" r="12.5" fill="#27BD68" />
                                    <path
                                        d="M7.5 12.8819L10.5769 15.9375L17.5 9.0625"
                                        stroke="white"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                    />
                                </svg>
                            </div>
                            <div class="time">0:08</div>
                        </div>

                        <div class="info_status">
                            <p id="shag">Шаг 1</p>
                            <p id="out">Отправленно в очередь</p>
                            <p id="res_stat_bar">Выполненно</p>
                        </div>
                    </div>

                    <div class="errors">
                        <div class="icon_and_time">
                            <div class="icon">
                                <svg
                                    width="30"
                                    height="30"
                                    viewBox="0 0 25 25"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <circle cx="12.5" cy="12.5" r="12.5" fill="#F6555F" />
                                    <path
                                        d="M17.6251 7.25L7.25 17.6251M7.25002 7.25L17.6251 17.6251"
                                        stroke="white"
                                        stroke-width="1.5"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                    />
                                </svg>
                            </div>
                            <div class="time">1:08</div>
                        </div>

                        <div class="info_status">
                            <p id="shag">Шаг 2</p>
                            <p id="out">Отправленно в очередь</p>
                            <p id="res_stat_bar_errors">Ошибка</p>
                        </div>
                    </div>

                    <div class="works">
                        <div class="icon_and_time">
                            <div class="icon">
                                <svg
                                    width="30"
                                    height="30"
                                    viewBox="0 0 26 26"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <circle
                                        cx="13"
                                        cy="13"
                                        r="10.7069"
                                        fill="#556FF6"
                                        stroke="#556FF6"
                                    />
                                    <circle cx="13" cy="13" r="12.5" stroke="#556FF6" />
                                    <path
                                        d="M13 18C15.7614 18 18 15.7614 18 13C18 10.2386 15.7614 8 13 8C10.2386 8 8 10.2386 8 13"
                                        stroke="white"
                                        stroke-linecap="round"
                                    />
                                </svg>
                            </div>
                            <div class="time">0:51</div>
                        </div>

                        <div class="info_status">
                            <p id="shag">Шаг 3</p>
                            <p id="out">Отправленно в очередь</p>
                            <p id="res_stat_bar_works">В работе</p>
                        </div>
                    </div>

                    <div class="cancellation">
                        <div class="icon_and_time">
                            <div class="icon">
                                <svg
                                    width="30"
                                    height="30"
                                    viewBox="0 0 25 25"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <circle cx="12.5" cy="12.5" r="12.5" fill="#F6AC55" />
                                    <path
                                        d="M17.6251 7.25L7.25 17.6251M7.25002 7.25L17.6251 17.6251"
                                        stroke="white"
                                        stroke-width="1.5"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                    />
                                </svg>
                            </div>
                            <div class="time">0:23</div>
                        </div>

                        <div class="info_status">
                            <p id="shag">Шаг 4</p>
                            <p id="out">Отправленно в очередь</p>
                            <p id="res_stat_bar_cancellation">Отменен</p>
                        </div>
                    </div>

                    <div class="expectation">
                        <div class="icon_and_time">
                            <div class="icon">
                                <svg
                                    width="30"
                                    height="30"
                                    viewBox="0 0 25 25"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <circle cx="12.5" cy="12.5" r="12.5" fill="#9293AB" />
                                    <path
                                        d="M14.525 6.25H10.475C8.12502 6.25 7.94377 8.3625 9.21252 9.5125L15.7875 15.4875C17.0562 16.6375 16.875 18.75 14.525 18.75H10.475C8.12502 18.75 7.94377 16.6375 9.21252 15.4875L15.7875 9.5125C17.0562 8.3625 16.875 6.25 14.525 6.25Z"
                                        stroke="white"
                                        stroke-width="1.5"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                    />
                                </svg>
                            </div>
                        </div>
                        <div class="info_status">
                            <p id="shag">Шаг 5</p>
                            <p id="out">Чтение</p>
                            <p id="res_stat_bar_expectation">Ожидание</p>
                        </div>
                    </div>
                </div>
                <div class="block_ditail">
                    <div class="header">
                        <div class="process">
                            <RouterLink id="link" to="/Process-Email-List"
                                ><p>Процессы</p></RouterLink
                            >
                            <svg
                                width="12"
                                height="13"
                                viewBox="0 0 15 16"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    fill-rule="evenodd"
                                    clip-rule="evenodd"
                                    d="M4.70278 13.0472C4.43241 12.7769 4.43241 12.3385 4.70278 12.0682L8.82858 7.94229L4.70278 3.8164C4.43241 3.54603 4.43241 3.10775 4.70278 2.83738C4.97315 2.56701 5.41143 2.56701 5.6818 2.83738L10.2972 7.45278C10.4271 7.58256 10.5 7.75869 10.5 7.94229C10.5 8.12589 10.4271 8.30201 10.2972 8.4318L5.6818 13.0472C5.41143 13.3176 4.97315 13.3176 4.70278 13.0472Z"
                                    fill="#9293AB"
                                />
                            </svg>
                            <p>Рассылка писем</p>
                            <svg
                                width="12"
                                height="13"
                                viewBox="0 0 15 16"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    fill-rule="evenodd"
                                    clip-rule="evenodd"
                                    d="M4.70278 13.0472C4.43241 12.7769 4.43241 12.3385 4.70278 12.0682L8.82858 7.94229L4.70278 3.8164C4.43241 3.54603 4.43241 3.10775 4.70278 2.83738C4.97315 2.56701 5.41143 2.56701 5.6818 2.83738L10.2972 7.45278C10.4271 7.58256 10.5 7.75869 10.5 7.94229C10.5 8.12589 10.4271 8.30201 10.2972 8.4318L5.6818 13.0472C5.41143 13.3176 4.97315 13.3176 4.70278 13.0472Z"
                                    fill="#9293AB"
                                />
                            </svg>
                            <p>agt72.ru</p>
                        </div>
                    </div>

                    <div class="name_client">
                        <RouterLink to="/Process-Email">
                            <svg
                                width="32"
                                height="33"
                                viewBox="0 0 36 37"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    fill-rule="evenodd"
                                    clip-rule="evenodd"
                                    d="M24.7133 6.38667C25.3622 7.03554 25.3622 8.08759 24.7133 8.73645L14.8114 18.6385L24.7133 28.5407C25.3622 29.1896 25.3622 30.2414 24.7133 30.8903C24.0644 31.5392 23.0126 31.5392 22.3637 30.8903L11.2866 19.8134C10.975 19.5019 10.8 19.0792 10.8 18.6385C10.8 18.1979 10.975 17.7752 11.2866 17.4637L22.3637 6.38667C23.0126 5.73781 24.0644 5.73781 24.7133 6.38667Z"
                                    fill="#9293AB"
                                />
                            </svg>
                        </RouterLink>
                        <RouterLink id="link" to="/Pprogress-Email-Detail"
                            ><h2>agt72.ru</h2></RouterLink
                        >
                    </div>

                    <div class="detail">
                        <div class="hed_detail">
                            <h2>Детали</h2>
                            <button class="custon_btn">
                                <svg
                                    width="24"
                                    height="24"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    xmlns="http://www.w3.org/2000/svg"
                                >
                                    <path
                                        fill-rule="evenodd"
                                        clip-rule="evenodd"
                                        d="M20.0755 16.4755C19.6429 16.9081 18.9415 16.9081 18.5089 16.4755L11.9076 9.87423L5.30613 16.4755C4.87354 16.9081 4.1723 16.9081 3.7397 16.4755C3.30711 16.0429 3.30711 15.3417 3.7397 14.9091L11.1243 7.52439C11.332 7.31666 11.6138 7.19995 11.9076 7.19995C12.2013 7.19995 12.4831 7.31666 12.6908 7.52439L20.0755 14.9091C20.508 15.3417 20.508 16.0429 20.0755 16.4755Z"
                                        fill="#9293AB"
                                        fill-opacity="0.5"
                                    />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <RouterView />
</template>

<style scoped>
@import '../assets/css/PprogressEmailDetail.css';
</style>

<script></script>

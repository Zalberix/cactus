import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'dashboard',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        },
        {
            path: '/processes',
            name: 'processes',
            component: () => import('@/views/core/type_worker/TypesWorkerPage.vue')
        },
        {
            path: '/tokens',
            name: 'tokens',
            component: () => import('@/views/core/tokens/TokensPage.vue')
        },
        {
            path: '/email',
            name: 'email',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        },
        {
            path: '/telegram',
            name: 'telegram',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        },
        {
            path: '/sms',
            name: 'sms',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        },
        {
            path: '/push',
            name: 'push',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        },
        {
            path: '/oneC',
            name: 'oneC',
            component: () => import('@/views/core/dashboard/DashboardPage.vue')
        }
    ]
})

export default router

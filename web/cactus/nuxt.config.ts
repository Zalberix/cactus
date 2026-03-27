import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  ssr: false,
  modules: [
    'shadcn-nuxt',
    '@pinia/nuxt',
    '@nuxtjs/i18n',
    '@nuxtjs/color-mode',
  ],
  shadcn: {
    prefix: '',
    componentDir: './app/components/ui',
  },
  i18n: {
    locales: [
      { code: 'en', language: 'en-US', file: 'en.json' },
      { code: 'ru', language: 'ru-RU', file: 'ru.json' },
    ],
    defaultLocale: 'en',
    lazy: true,
    langDir: 'locales',
  },
  colorMode: {
    classSuffix: '',
  },
  runtimeConfig: {
    public: {
      apiBase: '/api/v1',
    },
  },
  vite: {
    plugins: [tailwindcss()],
  },
  spaLoadingTemplate: true,
})

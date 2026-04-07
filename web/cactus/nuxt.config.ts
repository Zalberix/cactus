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
  css: ['~/assets/css/tailwind.css'],
  runtimeConfig: {
    public: {
      apiBase: '/api/v1',
    },
  },
  devServer: {
    host: '127.0.0.1',
  },
  vite: {
    plugins: [tailwindcss()],
    server: {
      watch: {
        awaitWriteFinish: {
          stabilityThreshold: 300,
          pollInterval: 100,
        },
      },
      hmr: {
        protocol: 'ws',
        host: 'localhost',
        clientPort: 80,
      },
    },
  },
  spaLoadingTemplate: true,
})

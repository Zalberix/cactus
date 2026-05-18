import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  hooks: {
    'vite:extendConfig'(viteConfig) {
      const fallbackClientEntry = '#app/entry';
      const fallbackServerEntry = '#app/entry-spa';

      const input = viteConfig.build?.rollupOptions?.input;

      if (!input) {
        viteConfig.build = viteConfig.build || {};
        viteConfig.build.rollupOptions = viteConfig.build.rollupOptions || {};
        viteConfig.build.rollupOptions.input = {
          entry: fallbackClientEntry,
          server: fallbackServerEntry,
        };
        return;
      }

      if (typeof input !== 'string' && !Array.isArray(input)) {
        const normalizedInput = {
          ...input as Record<string, string>,
          entry: (input as Record<string, string>).entry || fallbackClientEntry,
          server: (input as Record<string, string>).server || fallbackServerEntry,
        };

        viteConfig.build = viteConfig.build || {};
        viteConfig.build.rollupOptions = viteConfig.build.rollupOptions || {};
        viteConfig.build.rollupOptions.input = normalizedInput;
      }
    },
  },
  devtools: {
    enabled: true,

    timeline: {
      enabled: true,
    },
  },
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
    defaultLocale: 'ru',
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
    optimizeDeps: {
      include: [
        '@tanstack/vue-table',
        '@vee-validate/zod',
        'vee-validate',
        'zod',
      ],
    },
    server: {
      watch: {
        awaitWriteFinish: {
          stabilityThreshold: 300,
          pollInterval: 100,
        },
      },
      hmr: {
        protocol: 'ws',
        host: '127.0.0.1',
        clientPort: 3000,
      },
    },
  },
  spaLoadingTemplate: true,
})

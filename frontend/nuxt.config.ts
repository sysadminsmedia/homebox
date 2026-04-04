import { defineNuxtConfig } from "nuxt/config";

const baseURL = process.env.NUXT_APP_BASE_URL || "/";

// https://v3.nuxtjs.org/api/configuration/nuxt.config
export default defineNuxtConfig({
  ssr: false,

  components: {
    dirs: [],
  },

  build: {
    transpile: ["vue-i18n"],
  },

  modules: [
    "@nuxtjs/tailwindcss",
    "@pinia/nuxt",
    "@vueuse/nuxt",
    "@vite-pwa/nuxt",
    "unplugin-icons/nuxt",
    "shadcn-nuxt",
    "@nuxt/eslint",
  ],

  eslint: {
    config: {},
  },

  // Runtime config for OpenTelemetry
  // Note: otelEnabled is determined automatically by querying the backend status endpoint.
  // When the backend has telemetry enabled, the frontend will automatically enable it.
  runtimeConfig: {
    public: {
      // OpenTelemetry configuration (can be overridden by environment variables)
      otelServiceName: process.env.NUXT_PUBLIC_OTEL_SERVICE_NAME || "homebox-frontend",
      otelServiceVersion: process.env.NUXT_PUBLIC_OTEL_SERVICE_VERSION || "1.0.0",
      otelSampleRate: process.env.NUXT_PUBLIC_OTEL_SAMPLE_RATE || "1.0",
      otelDebug: process.env.NUXT_PUBLIC_OTEL_DEBUG || "false",
    },
  },

  nitro: {
    devProxy: {
      [baseURL + "api"]: {
        target: "http://localhost:7745/api",
        ws: true,
        changeOrigin: true,
      },
    },
  },

  app: {
    baseURL,
    head: {
      script: [{ src: baseURL + "set-theme.js" }],
    },
  },

  css: ["@/assets/css/main.css"],

  pwa: {
    workbox: {
      navigateFallbackDenylist: [RegExp(`^${baseURL}/api`)],
      cleanupOutdatedCaches: true,
      runtimeCaching: [
        {
          urlPattern: RegExp(`^${baseURL}/api`),
          handler: "NetworkFirst",
          method: "GET",
          options: {
            cacheName: "api-cache",
            cacheableResponse: { statuses: [0, 200] },
            expiration: { maxAgeSeconds: 60 * 60 * 24 },
          },
        },
      ],
    },
    registerType: "autoUpdate",
    injectRegister: "script",
    injectManifest: {
      swSrc: "sw.js",
    },
    devOptions: {
      // Enable to troubleshoot during development
      enabled: false,
    },
    manifest: {
      name: "Homebox",
      short_name: "Homebox",
      description: "Home Inventory App",
      theme_color: "#5b7f67",
      start_url: baseURL.replace(/\/$/, "") + "/home",
      icons: [
        {
          src: "pwa-192x192.png",
          sizes: "192x192",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
          purpose: "any maskable",
        },
      ],
    },
  },
  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {},
    },
  },

  compatibilityDate: "2024-11-29",
});

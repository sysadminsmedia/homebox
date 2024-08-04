import { resolve, dirname } from "node:path";
import { fileURLToPath } from "url";
import { defineNuxtConfig } from "nuxt/config";
import VueI18nVitePlugin from "@intlify/unplugin-vue-i18n/vite";

// https://v3.nuxtjs.org/api/configuration/nuxt.config
export default defineNuxtConfig({
  ssr: false,
  build: {
    transpile: ["vue-i18n"],
  },
  vite: {
    plugins: [
      VueI18nVitePlugin({
        include: [resolve(dirname(fileURLToPath(import.meta.url)), "./locales/**")],
      }),
    ],
  },
  modules: [
    "@nuxtjs/tailwindcss",
    "@pinia/nuxt",
    "@vueuse/nuxt",
    "@vite-pwa/nuxt",
    "./nuxt.proxyoverride.ts",
    "unplugin-icons/nuxt",
  ],
  nitro: {
    devProxy: {
      "/api": {
        target: "http://localhost:7745/api",
        ws: true,
        changeOrigin: true,
      },
    },
  },
  css: ["@/assets/css/main.css"],
  pwa: {
    workbox: {
      navigateFallbackDenylist: [/^\/api/],
    },
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
      start_url: "/home",
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
});

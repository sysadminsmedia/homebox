import { defineConfig } from 'vitepress'
import enMenu from "./menus/en.mjs";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  ignoreDeadLinks: [
    /^https?:\/\/localhost:7745/,
  ],

  title: "HomeBox",
  description: "A simple home inventory management software",
  lastUpdated: true,
  sitemap: {
    hostname: 'https://homebox.software',
  },

  head: [
    ['link', { rel: 'icon', href: '/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
    ['meta', { name: 'og:title', content: 'HomeBox' }],
    ['meta', { name: 'og:description', content: 'A simple home inventory management software' }],
    ['meta', { name: 'og:image', content: '/homebox-email-banner.jpg' }],
    ['meta', { name: 'twitter:card', content: 'summary' }],
  ],

  locales: {
    en: {
      label: 'English',
      lang: 'en',
    }
  },

  themeConfig: {
    logo: '/lilbox.svg',

    search: {
      provider: 'local'
    },
    editLink: {
      pattern: 'https://github.com/sysadminsmedia/homebox/edit/main/docs/:path'
    },
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'API', link: 'https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/sysadminsmedia/homebox/main/docs/docs/api/openapi-2.0.json' },
      { text: 'Demo', link: 'https://demo.homebox.software' },
    ],

    sidebar: {
        '/en/': enMenu,
    },

    socialLinks: [
      { icon: 'discord', link: 'https://discord.homebox.software' },
      { icon: 'github', link: 'https://git.homebox.software' },
      { icon: 'mastodon', link: 'https://noc.social/@sysadminszone' },
    ]
  }
})

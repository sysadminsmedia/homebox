import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "HomeBox",
  description: "A simple home inventory management software",
  lastUpdated: true,
  sitemap: {
    hostname: 'https://homebox.sysadminsmedia.com',
  },

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
      { text: 'Home', link: '/' },
      { text: 'API', link: 'https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/sysadminsmedia/homebox/main/docs/docs/api/openapi-2.0.json' }
    ],

    sidebar: {
      '/en/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Quick Start', link: '/en/quick-start' },
            { text: 'Tips and Tricks', link: '/en/tips-tricks' }
          ]
        },
        {
          text: 'Advanced',
          items: [
            { text: 'Import CSV', link: '/en/import-csv' },
          ]
        },
        {
          text: 'Contributing',
          items: [
            { text: 'Get Started', link: '/en/contribute/get-started' },
            { text: 'Bounty Program', link: '/en/contribute/bounty' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'discord', link: 'https://discord.gg/aY4DCkpNA9' },
      { icon: 'github', link: 'https://github.com/sysadminsmedia/homebox' },
      { icon: 'mastodon', link: 'https://noc.social/@sysadminszone' },
    ]
  }
})

import { defineConfig } from 'vitepress'

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
      '/en/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Quick Start', link: '/en/quick-start' },
            { text: 'Installation', link: '/en/installation' },
            { text: 'Configure Homebox', link: '/en/configure-homebox' },
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
      { icon: 'discord', link: 'https://discord.homebox.software' },
      { icon: 'github', link: 'https://git.homebox.software' },
      { icon: 'mastodon', link: 'https://noc.social/@sysadminszone' },
    ]
  }
})

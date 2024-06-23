import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "HomeBox",
  description: "A simple home inventory management software",
  lastUpdated: true,
  sitemap: {
    hostname: 'https://homebox.sysadminsmedia.com',
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

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Quick Start', link: '/quick-start' },
          { text: 'Tips and Tricks', link: '/tips-tricks' }
        ]
      },
      {
        text: 'Advanced',
        items: [
          { text: 'Import CSV', link: '/import-csv' },
          { text: 'Build from Source', link: '/build' }
        ]
      },
    ],

    socialLinks: [
      { icon: 'discord', link: 'https://discord.gg/aY4DCkpNA9' },
      { icon: 'github', link: 'https://github.com/sysadminsmedia/homebox' },
      { icon: 'mastodon', link: 'https://noc.social/@sysadminszone' },
    ]
  }
})

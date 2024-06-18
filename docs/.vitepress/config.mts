import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "HomeBox",
  description: "A simple home inventory management software",
  lastUpdated: true,
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
      { icon: 'github', link: 'https://github.com/sysadminsmedia/homebox' }
    ]
  }
})

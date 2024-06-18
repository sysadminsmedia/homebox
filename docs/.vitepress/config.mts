import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "HomeBox",
  description: "A simple home inventory management software",
  themeConfig: {
    logo: '/lilbox.svg',
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
          { item: 'Import CSV', link: '/import-csv' },
          { item: 'Build', link: '/build'}
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/sysadminsmedia/homebox' }
    ]
  }
})

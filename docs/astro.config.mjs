// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeNova from 'starlight-theme-nova';
import starlightChangelogs, {
  makeChangelogsSidebarLinks,
} from 'starlight-changelogs';
import starlightOpenAPI, { openAPISidebarGroups } from 'starlight-openapi';
import starlightGitHubAlerts from 'starlight-github-alerts';
import starlightAutoDrafts from 'starlight-auto-drafts'

import icon from 'astro-icon';
import starlightSidebarTopics from 'starlight-sidebar-topics';
import cloudflare from '@astrojs/cloudflare';

import tailwindcss from '@tailwindcss/vite';

// https://astro.build/config
export default defineConfig({
  experimental: {
    svgo: true,
    contentIntellisense: true,
    clientPrerender: true,
    chromeDevtoolsWorkspace: true,
  },

  prefetch: {
    prefetchAll: true,
    defaultStrategy: 'hover',
  },

    site: 'https://homebox.software',

    integrations: [starlight({
        components: {
            SocialIcons: './src/components/SocialIcon.astro',
        },
        logo: {
            src: './src/assets/lilbox.svg',
        },
        favicon: './src/assets/favicon.svg',
        editLink: {
            baseUrl: 'https://github.com/sysadminsmedia/homebox/edit/main/docs/'
        },
        lastUpdated: true,
        plugins: [
            starlightThemeNova({
                nav: [
                    {label: 'Demos', href: 'https://demo.homebox.software'},
                    {label: 'API Docs', href: '/api'},
                ]
            }),
            starlightGitHubAlerts(),
            starlightChangelogs(),
            starlightAutoDrafts(),
            starlightOpenAPI([{
                base: 'api',
                schema: 'https://raw.githubusercontent.com/sysadminsmedia/homebox/refs/heads/main/docs/en/api/openapi-3.0.json',
            }])
        ],
        title: 'Homebox',
        defaultLocale: 'en',
        locales: {
            en: {
                label: 'English',
            },
            de: {
                label: 'Deutsch',
            }
        },
        sidebar: [
            {
                label: 'Getting Started',
                items: [
                    {label: 'Quick Start', slug: 'quick-start'},
                    {label: 'Install', slug: 'quick-start/install'},
                    {label: 'Configure', autogenerate: {directory: 'quick-start/configure'}}
                ]
            },
            {
                label: 'User Guide',
                autogenerate: {directory: 'user-guide'},
            },
            {
                label: 'Advanced',
                autogenerate: {directory: 'advanced'}
            },
            {
                label: 'Contributing',
                items: [
                    {label: 'Getting Started', slug: 'contribute/getting-started'},
                    {label: 'Bounty Program', slug: 'contribute/bounty-program'},
                    {label: 'Development', autogenerate: {directory: 'contribute/development'}},
                    {label: 'Translations', autogenerate: {directory: 'contribute/translate'}},
                    {label: 'Documentation', autogenerate: {directory: 'contribute/documentation'}}
                ]
            },
            ...makeChangelogsSidebarLinks([{
                label: 'Changelogs',
                type: 'all',
                base: 'changelog'
            }]),
            ...openAPISidebarGroups,
        ],
    }), icon()],

  adapter: cloudflare(),

  vite: {
    plugins: [tailwindcss()],
  },
});
// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeNova from 'starlight-theme-nova';
import starlightChangelogs, { makeChangelogsSidebarLinks } from 'starlight-changelogs';
import starlightOpenAPI, { openAPISidebarGroups } from 'starlight-openapi';
import starlightGitHubAlerts from 'starlight-github-alerts';
import icon from 'astro-icon';
import starlightSidebarTopics from 'starlight-sidebar-topics';
import starlightAutoDrafts from "starlight-auto-drafts";
import cloudflare from '@astrojs/cloudflare';
import tailwindcss from "@tailwindcss/vite";

// https://astro.build/config
// @ts-ignore
export default defineConfig({
    experimental: {
        svgo: true,
        contentIntellisense: true,
        clientPrerender: true,
        chromeDevtoolsWorkspace: true,
        /*csp: {
            algorithm: 'SHA-384',
            directives: [
                "img-src 'self' data: https://translate.sysadminsmedia.com;"
            ]
        } Turn this off for now while we work on things, sort it later*/
    },

    prefetch: {
        prefetchAll: true,
        defaultStrategy: 'hover',
    },

    site: 'https://homebox.software',
    integrations: [
        starlight({
            components: {
                SocialIcons: './src/components/theme/SocialIcon.astro',
                SiteTitle: './src/components/theme/SiteTitle.astro',
                Header: './src/components/theme/Header.astro',
                PageFrame: './src/components/theme/PageFrame.astro',
                Pagination: './src/components/theme/Pagination.astro',
                MobileMenuToggle: './src/components/theme/MobileMenuToggle.astro',
                Search: './src/components/theme/Search.astro',
                ThemeSelect: './src/components/theme/ThemeSelect.astro',
                TwoColumnContent: './src/components/theme/TwoColumnContent.astro',
                Hero: './src/components/theme/Hero.astro',
            },
            logo: {
                src: './src/assets/lilbox.svg',
            },
            favicon: './src/assets/favicon.svg',
            editLink: {
                baseUrl: 'https://github.com/sysadminsmedia/homebox/edit/main/docs/',
            },
            lastUpdated: true,
            plugins: [
                starlightThemeNova({
                    // nav: [
                    //     { label: 'Demos', href: 'https://demo.homebox.software' },
                    //     { label: 'API Docs', href: '/api' },
                    // ],
                }),
                starlightGitHubAlerts(),
                starlightChangelogs(),
                starlightAutoDrafts(),
                starlightOpenAPI([
                    {
                        base: 'en/api',
                        schema:
                            'https://raw.githubusercontent.com/sysadminsmedia/homebox/refs/heads/main/docs/en/api/openapi-3.0.json',
                    },
                ]),
                starlightSidebarTopics(
                    [
                        {
                            label: 'Documentation',
                            link: '/quick-start/',
                            icon: 'open-book',
                            items: [
                                {
                                    label: 'Getting Started',
                                    items: [
                                        {
                                            label: 'Quick Start',
                                            slug: 'quick-start',
                                        },
                                        {
                                            label: 'Install',
                                            slug: 'quick-start/install',
                                        },
                                        {
                                            label: 'Configure',
                                            autogenerate: { directory: 'quick-start/configure' },
                                        },
                                    ],
                                },
                                {
                                    label: 'User Guide',
                                    autogenerate: { directory: 'user-guide' },
                                },
                                {
                                    label: 'Advanced',
                                    collapsed: true,
                                    autogenerate: { directory: 'advanced' },
                                },
                                {
                                    label: 'Contributing',
                                    collapsed: true,
                                    items: [
                                        {
                                            label: 'Getting Started',
                                            slug: 'contribute/getting-started',
                                        },
                                        {
                                            label: 'Bounty Program',
                                            slug: 'contribute/bounty-program',
                                        },
                                        {
                                            label: 'Code of Conduct',
                                            slug: 'contribute/code-of-conduct',
                                        },
                                        {
                                            label: 'Development',
                                            autogenerate: { directory: 'contribute/development' },
                                        },
                                        {
                                            label: 'Translations',
                                            autogenerate: { directory: 'contribute/translate' },
                                        },
                                        {
                                            label: 'Documentation',
                                            autogenerate: { directory: 'contribute/documentation' },
                                        },
                                    ],
                                },
                                {
                                    label: 'Analytics',
                                    items: [
                                        {
                                            label: 'Purpose & Data',
                                            slug: 'analytics',
                                        },
                                        {
                                            label: 'Privacy Policy',
                                            slug: 'analytics/privacy',
                                        }
                                    ]
                                }
                            ],
                        },
                        {
                            label: 'Changelogs',
                            link: '/changelog/',
                            icon: 'information',
                            items: makeChangelogsSidebarLinks([
                                {
                                    label: 'Changelogs',
                                    type: 'all',
                                    base: 'changelog',
                                },
                            ]),
                        },
                        {
                            label: 'API',
                            // TODO: the api link is broken bc this links to /en/api/ not /api/
                            link: '/api/',
                            icon: 'forward-slash',
                            items: openAPISidebarGroups,
                        },
                        {
                            label: 'Demo',
                            link: 'https://demo.homebox.software',
                            icon: 'puzzle',
                        },
                        {
                            label: 'Blog',
                            link: 'https://blog.homebox.software',
                            icon: 'document',
                        },
                    ],
                    {
                        exclude: ['**/*'],
                    }
                ),
            ],
            title: 'Homebox',
            defaultLocale: 'en',
            locales: {
                en: {
                    label: 'English',
                },
                de: {
                    label: 'Deutsch',
                },
            },
            customCss: [
                './src/styles/global.css',
            ],
            expressiveCode: {
                removeUnusedThemes: true, // Try to reduce bundle size by removing unused themes
                useThemedSelectionColors: true,
                minSyntaxHighlightingColorContrast: 5.0, // Ensure minimum contrast is at least .5 units higher than WCAG 2.2 AA
                shiki: { // We set the languages we actually use to try to reduce bundle sizes
                    bundledLangs: ['bash', 'typescript', 'javascript', 'json', 'yaml', 'go', 'systemd', 'vue', 'vue-html', 'astro', 'css', 'sql'],
                }
            }
        }),
        icon({
            include: { // Specify which icons to include in the final bundle (reduce bundle size)
                'material-symbols': ['home-work', 'edit-document', 'family-group', 'settings', 'book-5', 'flash-on', 'lock', 'package', 'folder-open', 'label'],
                'simple-icons': ['discord', 'github', 'lemmy', 'reddit', 'mastodon'],
            }
        }),
    ],

    //adapter: cloudflare(),

    vite: {
        plugins: [tailwindcss()],
    },
});
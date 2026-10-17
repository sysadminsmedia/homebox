import { defineCollection } from 'astro:content';
import { docsLoader, i18nLoader } from '@astrojs/starlight/loaders';
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema';
import { changelogsLoader } from 'starlight-changelogs/loader'

export const collections = {
	docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
	i18n: defineCollection({ loader: i18nLoader(), schema: i18nSchema() }),
	changelogs: defineCollection({
		loader: changelogsLoader([
			{
				base: 'changelog',
				provider: 'github',
				repo: 'homebox',
				owner: 'sysadminsmedia',
				// Unauthenticated GitHub API requests are rate limited per IP, which shared CI
				// builders (e.g. Cloudflare Pages) exhaust quickly. Set GITHUB_TOKEN in the build
				// environment to use the authenticated limit instead.
				token: process.env.GITHUB_TOKEN,
			}
		]),
	}),
};

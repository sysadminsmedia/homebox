import { defineCollection } from 'astro:content';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';
import { changelogsLoader } from 'starlight-changelogs/loader'

// Releases are fetched from the GitHub API at build time. Unauthenticated
// requests are capped at 60/hour per IP, which shared CI runners (Cloudflare
// Pages) exhaust routinely and the build then fails outright. A token raises
// that to 5,000/hour. Optional: builds without one still work, they just fall
// back to the shared limit.
const githubToken = process.env.GITHUB_TOKEN || undefined;

export const collections = {
	docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
	changelogs: defineCollection({
		loader: changelogsLoader([
			{
				base: 'changelog',
				provider: 'github',
				repo: 'homebox',
				owner: 'sysadminsmedia',
				token: githubToken
			}
		]),
	}),
};

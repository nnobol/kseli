import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const outputDir = process.env.CLIENT_OUTPUT_DIR || '../builds/client-new';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter(
			{
				pages: outputDir,
				assets: outputDir,
				fallback: 'index.html',
				precompress: true,
				strict: false
			}
		)
	}
};

export default config;

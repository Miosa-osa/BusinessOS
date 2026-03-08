import adapterNode from '@sveltejs/adapter-node';
import adapterStatic from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

// Determine which adapter to use based on environment
const isElectronBuild = process.env.ELECTRON_BUILD === 'true';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Use adapter-static for Electron builds, adapter-node for Cloud Run
		adapter: isElectronBuild
			? adapterStatic({
					pages: 'build',
					assets: 'build',
					fallback: 'index.html', // SPA fallback for client-side routing
					precompress: false,
					strict: false
				})
			: adapterNode(),
		// For Electron builds, disable all prerendering (pure SPA)
		prerender: isElectronBuild
			? { entries: [] }
			: undefined,
		// For Electron builds, use relative paths (file:// protocol compatibility)
		paths: isElectronBuild
			? { relative: true }
			: undefined
	}
};

export default config;

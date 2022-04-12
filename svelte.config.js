import adapter from '@sveltejs/adapter-static';
import { isoImport } from 'vite-plugin-iso-import';
import { defineConfig } from 'vite';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false
		}),
		vite: defineConfig({
			plugins: [isoImport()]
		})
	}
};

export default config;

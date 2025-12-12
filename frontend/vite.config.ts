import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	server: {
		proxy: {
			// Proxy specific backend API routes to FastAPI
			// Auth routes (/api/auth/*) are handled by Better Auth via SvelteKit hooks
			'/api/chat': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/projects': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/contexts': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/team': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/dashboard': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/mcp': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/daily': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/settings': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/artifacts': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/nodes': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/clients': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/api/deals': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
			'/health': {
				target: 'http://localhost:8000',
				changeOrigin: true,
			},
		}
	}
});

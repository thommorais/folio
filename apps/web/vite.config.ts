import babel from '@rolldown/plugin-babel';
import tailwindcss from '@tailwindcss/vite';
import tanstackRouter from '@tanstack/router-plugin/vite';
import react, { reactCompilerPreset } from '@vitejs/plugin-react';
import { resolve } from 'node:path';
import { intlayer } from 'vite-intlayer';
import { comlink } from 'vite-plugin-comlink';
import { VitePWA } from 'vite-plugin-pwa';
import { defineConfig } from 'vitest/config';

// oxlint-disable-next-line import/no-default-export
export default defineConfig({
	// host exposes the dev server on the LAN so a phone can reach it; basicSsl
	// below serves it over HTTPS. The camera needs a secure context, and iOS
	// Safari has no "treat this LAN address as secure" escape hatch the way
	// Chrome does, so without both the scanner cannot be tested on a phone at
	// all. The certificate is self-signed and accepted once per device.
	server: {
		host: true,
	},
	build: {
		sourcemap: true,
		rolldownOptions: {
			output: {
				// Split the stable framework core into its own chunk so it stays
				// cached across app deploys (route/feature chunks are already
				// code-split on demand by the TanStack Router plugin above).
				advancedChunks: {
					groups: [
						{
							name: 'react-vendor',
							test: /[\\/]node_modules[\\/](react|react-dom|scheduler|@tanstack[\\/]react-router)[\\/]/,
						},
					],
				},
			},
		},
	},
	plugins: [
		tanstackRouter({
			routesDirectory: './src/routes',
			generatedRouteTree: './src/routeTree.gen.ts',
			autoCodeSplitting: true,
			target: 'react',
		}),
		react(),
		babel({ presets: [reactCompilerPreset({
  logger: {
    logEvent(filename, event) {
      switch (event.kind) {
        case 'CompileSuccess': {
          console.log(`✅ Compiled: ${filename}`);
          break;
        }
        case 'CompileError': {
          console.log(`❌ Skipped: ${filename}`);
          break;
        }
        default: {}
      }
    }
  }
})] }),
		tailwindcss(),
		intlayer(),
		comlink(),
		VitePWA({
			registerType: 'prompt',
			// vite-plugin-pwa 1.3.0 resolves its virtual register module against a
			// __dirname that leaks in from esbuild under Vite 8, so the injected
			// client file is looked for inside esbuild and the build dies. The app
			// registers the worker itself in register-sw.ts instead.
			injectRegister: null,
			includeAssets: ['favicon.svg', 'icons/favicon.ico', 'icons/apple-touch-icon-180x180.png'],
			manifest: {
				id: '/',
				name: 'folio',
				short_name: 'folio',
				description: 'The shared channel between developers and their coding agents.',
				lang: 'en',
				dir: 'ltr',
				start_url: '/',
				scope: '/',
				display: 'standalone',
				display_override: ['window-controls-overlay', 'standalone', 'minimal-ui'],
				background_color: '#060E0B',
				theme_color: '#060E0B',
				categories: ['productivity', 'developer'],
				icons: [
					{ src: '/icons/pwa-64x64.png', sizes: '64x64', type: 'image/png', purpose: 'any' },
					{ src: '/icons/pwa-192x192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
					{ src: '/icons/pwa-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
					{ src: '/icons/maskable-icon-512x512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
				],
				screenshots: [
					{
						src: '/screenshots/wide-1280x800.png',
						sizes: '1280x800',
						type: 'image/png',
						form_factor: 'wide',
						label: 'folio',
					},
					{
						src: '/screenshots/narrow-720x1280.png',
						sizes: '720x1280',
						type: 'image/png',
						form_factor: 'narrow',
						label: 'folio',
					},
				],
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,svg,png,ico,woff2}'],
				// A 2.24 MB rapier chunk arrives transitively through @thom/ui and no
				// source file imports it. Precaching it would cost every install that
				// much for code the app never runs, so it is left to the network.
				// Raising maximumFileSizeToCacheInBytes instead would ship it to
				// everyone; dropping the dependency is a separate change.
				globIgnores: ['**/rapier-*.js'],
				// The SPA falls back to index.html, which would swallow a missing
				// API route and answer it with the app shell instead of a 404.
				navigateFallbackDenylist: [/^\/api/, /^\/_/],
				cleanupOutdatedCaches: true,
				runtimeCaching: [
					{
						// Google Fonts ships immutable, hashed files, so they are worth
						// keeping across deploys rather than refetching on every visit.
						urlPattern: /^https:\/\/fonts\.(googleapis|gstatic)\.com\//,
						handler: 'CacheFirst',
						options: {
							cacheName: 'google-fonts',
							expiration: { maxEntries: 20, maxAgeSeconds: 60 * 60 * 24 * 365 },
							cacheableResponse: { statuses: [0, 200] },
						},
					},
				],
			},
			devOptions: { enabled: false },
		}),
	],
	test: {
		environment: 'happy-dom',
		setupFiles: ['./src/test/setup-env.ts'],
	},
	worker: {
		plugins: () => [comlink()],
	},
	resolve: {
		alias: {
			_: resolve(import.meta.dirname, './src'),
			'next/image': resolve(import.meta.dirname, './src/libs/next-image-shim.tsx'),
			'next/link': resolve(import.meta.dirname, './src/libs/next-link-shim.tsx'),
			'next/navigation': resolve(import.meta.dirname, './src/libs/next-navigation-shim.ts'),
		},
	},
})

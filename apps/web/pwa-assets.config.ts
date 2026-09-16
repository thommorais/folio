import { defineConfig } from '@vite-pwa/assets-generator/config'

// Run `pnpm generate-pwa-assets` after changing the source SVG.
// Paddings are spelled out because minimal2023Preset's 0.3 default assumes a
// square mark and crops this one. https://web.dev/articles/maskable-icon
// oxlint-disable-next-line import/no-default-export
export default defineConfig({
	headLinkOptions: { preset: '2023' },
	preset: {
		transparent: {
			sizes: [64, 192, 512],
			favicons: [[48, 'favicon.ico']],
		},
		maskable: {
			sizes: [512],
			padding: 0.4,
			resizeOptions: { fit: 'contain', background: '#F5F4ED' },
		},
		apple: {
			sizes: [180],
			padding: 0.28,
			resizeOptions: { fit: 'contain', background: '#F5F4ED' },
		},
	},
	images: ['public/icons/icon.svg'],
})

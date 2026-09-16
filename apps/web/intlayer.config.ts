import { type IntlayerConfig, Locales } from 'intlayer'

const config: IntlayerConfig = {
	internationalization: {
		locales: [Locales.ENGLISH, Locales.PORTUGUESE],
		defaultLocale: Locales.PORTUGUESE,
	},
}

// oxlint-disable-next-line import/no-default-export -- intlayer config
export default config

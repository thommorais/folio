/// <reference types="vite/client" />
/// <reference types="vite-plugin-comlink/client" />
/// <reference types="vite-plugin-pwa/react" />

declare module '@fontsource-variable/jost'
declare module '@fontsource-variable/geist'

interface ImportMetaEnv {
	/** Base URL of the @thom/notify web-push service (e.g. https://notify.thom.place). */
	readonly VITE_NOTIFY_URL?: string
	readonly VITE_API_URL?: string
	readonly VITE_WEBAPP_URL?: string
	readonly VITE_SITE_NAME?: string
	readonly VITE_VAPID_PUBLIC_KEY?: string
	readonly VITE_VAPID_PRIVATE_KEY?: string
}

interface ImportMeta {
	readonly env: ImportMetaEnv
}

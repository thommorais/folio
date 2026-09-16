// Registered by hand rather than through virtual:pwa-register: that module is
// generated from a file vite-plugin-pwa cannot locate under Vite 8. The
// contract here is the same, an onNeedRefresh callback plus an update trigger.
const SW_URL = '/sw.js'

export type ServiceWorkerHandle = {
	readonly update: () => Promise<void>
}

export const registerServiceWorker = (onNeedRefresh: (handle: ServiceWorkerHandle) => void): void => {
	if (!('serviceWorker' in navigator)) return

	const activate = (waiting: ServiceWorker) => {
		onNeedRefresh({
			update: async () => {
				// The worker calls skipWaiting on this message; reloading before it
				// activates would just serve the old shell again.
				waiting.postMessage({ type: 'SKIP_WAITING' })
				await new Promise<void>(resolve => {
					navigator.serviceWorker.addEventListener('controllerchange', () => resolve(), { once: true })
				})
				location.reload()
			},
		})
	}

	void navigator.serviceWorker.register(SW_URL, { scope: '/' }).then(registration => {
		if (registration.waiting) {
			activate(registration.waiting)
		}

		registration.addEventListener('updatefound', () => {
			const installing = registration.installing
			if (!installing) return

			installing.addEventListener('statechange', () => {
				// Without a controller this is the first install, which is already
				// the newest version: prompting there would be a reload to nowhere.
				if (installing.state === 'installed' && navigator.serviceWorker.controller) {
					activate(installing)
				}
			})
		})
	})
}

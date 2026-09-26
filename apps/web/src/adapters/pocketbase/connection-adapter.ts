import type { ConnectionPort } from '_/core/ports/connection'
import { getPocketBaseClient } from './client'

type ConnectTopic = {
	readonly subscribe: (topic: string, handler: () => void) => Promise<unknown>
	onDisconnect?: (activeSubscriptions: string[]) => void
}

export const createConnectionAdapter = (realtime?: ConnectTopic): ConnectionPort => {
	const source = realtime ?? getPocketBaseClient().realtime
	const listeners = new Set<() => void>()
	const statusListeners = new Set<(online: boolean) => void>()
	let connected = false

	const notify = () => {
		for (const listener of listeners) listener()
	}

	const setOnline = (online: boolean) => {
		for (const listener of statusListeners) listener(online)
	}

	source.onDisconnect = activeSubscriptions => {
		if (activeSubscriptions.length > 0) setOnline(false)
	}

	void source.subscribe('PB_CONNECT', () => {
		setOnline(true)

		if (!connected) {
			connected = true
			return
		}

		notify()
	})

	return {
		onReconnect: listener => {
			listeners.add(listener)

			return () => {
				listeners.delete(listener)
			}
		},
		onStatusChange: listener => {
			statusListeners.add(listener)

			return () => {
				statusListeners.delete(listener)
			}
		},
		retry: notify,
	}
}

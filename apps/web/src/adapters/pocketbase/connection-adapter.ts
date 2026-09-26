import type { ConnectionPort } from '_/core/ports/connection'
import { getPocketBaseClient } from './client'

type ConnectTopic = {
	readonly subscribe: (topic: string, handler: () => void) => Promise<unknown>
}

export const createConnectionAdapter = (realtime?: ConnectTopic): ConnectionPort => {
	const source = realtime ?? getPocketBaseClient().realtime
	const listeners = new Set<() => void>()
	let connected = false

	const notify = () => {
		for (const listener of listeners) listener()
	}

	void source.subscribe('PB_CONNECT', () => {
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
		retry: notify,
	}
}

import type { ConnectionPort } from '_/core/ports/connection'
import { getPocketBaseClient } from './client'

type ConnectTopic = {
	readonly subscribe: (topic: string, handler: () => void) => Promise<unknown>
}

export const createConnectionAdapter = (realtime?: ConnectTopic): ConnectionPort => {
	const source = realtime ?? getPocketBaseClient().realtime
	const listeners = new Set<() => void>()
	let connected = false

	void source.subscribe('PB_CONNECT', () => {
		if (!connected) {
			connected = true
			return
		}

		for (const listener of listeners) listener()
	})

	return {
		onReconnect: listener => {
			listeners.add(listener)

			return () => {
				listeners.delete(listener)
			}
		},
	}
}

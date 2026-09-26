import { describe, expect, it, vi } from 'vitest'
import { createConnectionAdapter } from './connection-adapter'

type Handler = () => void

const fakeRealtime = () => {
	const handlers: Handler[] = []

	const realtime = {
		onDisconnect: undefined as ((active: string[]) => void) | undefined,
		subscribe: vi.fn(async (_topic: string, handler: Handler) => {
			handlers.push(handler)
			return async () => {}
		}),
		connect: () => {
			for (const handler of handlers) handler()
		},
		drop: (active: string[] = ['issues']) => realtime.onDisconnect?.(active),
	}

	return realtime
}

describe('createConnectionAdapter', () => {
	it('stays quiet on the first connect', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		adapter.onReconnect(listener)

		realtime.connect()

		expect(listener).not.toHaveBeenCalled()
	})

	it('fires on every connect after the first', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		adapter.onReconnect(listener)

		realtime.connect()
		realtime.connect()
		realtime.connect()

		expect(listener).toHaveBeenCalledTimes(2)
	})

	it('tells a listener that arrived after the first connect about the next one', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		realtime.connect()

		const listener = vi.fn()
		adapter.onReconnect(listener)
		realtime.connect()

		expect(listener).toHaveBeenCalledTimes(1)
	})

	it('notifies every listener', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const first = vi.fn()
		const second = vi.fn()
		adapter.onReconnect(first)
		adapter.onReconnect(second)

		realtime.connect()
		realtime.connect()

		expect(first).toHaveBeenCalledTimes(1)
		expect(second).toHaveBeenCalledTimes(1)
	})

	it('stops notifying a listener that removed itself', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		const remove = adapter.onReconnect(listener)

		realtime.connect()
		remove()
		realtime.connect()

		expect(listener).not.toHaveBeenCalled()
	})

	it('subscribes to the connect topic once however many listeners attach', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		adapter.onReconnect(vi.fn())
		adapter.onReconnect(vi.fn())

		expect(realtime.subscribe).toHaveBeenCalledTimes(1)
		expect(realtime.subscribe).toHaveBeenCalledWith('PB_CONNECT', expect.any(Function))
	})

	it('retries on demand, telling every listener to reload', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const first = vi.fn()
		const second = vi.fn()
		adapter.onReconnect(first)
		adapter.onReconnect(second)

		adapter.retry()

		expect(first).toHaveBeenCalledTimes(1)
		expect(second).toHaveBeenCalledTimes(1)
	})

	it('reports going offline when the connection drops with live subscriptions', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		adapter.onStatusChange(listener)
		realtime.connect()

		realtime.drop()

		expect(listener).toHaveBeenLastCalledWith(false)
	})

	it('stays online when the last subscription is closed on purpose', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		adapter.onStatusChange(listener)
		realtime.connect()

		realtime.drop([])

		expect(listener).not.toHaveBeenCalledWith(false)
	})

	it('reports coming back online on reconnect', () => {
		const realtime = fakeRealtime()
		const adapter = createConnectionAdapter(realtime)
		const listener = vi.fn()
		adapter.onStatusChange(listener)
		realtime.connect()
		realtime.drop()

		realtime.connect()

		expect(listener).toHaveBeenLastCalledWith(true)
	})
})

import { act, renderHook, waitFor } from '@testing-library/react'
import type { ConnectionPort } from '_/core/ports/connection'
import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'
import { describe, expect, it, vi } from 'vitest'
import { useLiveRecord } from './use-live-record'
import { Status } from '_/lib/async-status'

type Row = { readonly id: string; readonly slug: string; readonly title: string }

const row = (id: string, title = id, slug = id): Row => ({ id, slug, title })

const fakeConnection = () => {
	const listeners = new Set<() => void>()

	return {
		port: {
			onReconnect: (listener: () => void) => {
				listeners.add(listener)
				return () => listeners.delete(listener)
			},
		} satisfies ConnectionPort,
		reconnect: () => {
			for (const listener of listeners) listener()
		},
	}
}

const setup = (initial: Row) => {
	let onChange: (row: Row) => void = () => {}
	let onGone: () => void = () => {}
	const close = vi.fn(async () => {})
	let current: Result<Row> = ok(initial)

	return {
		load: vi.fn(async () => current),
		subscribe: vi.fn(async (_id: string, change: (row: Row) => void, gone: () => void) => {
			onChange = change
			onGone = gone
			return ok(close as Unsubscribe)
		}),
		close,
		change: (next: Row) => act(() => onChange(next)),
		remove: () => act(() => onGone()),
		setServer: (next: Result<Row>) => {
			current = next
		},
	}
}

describe('useLiveRecord', () => {
	it('loads the record', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, data: row('a') }))
	})

	it('subscribes by the id the load returned, not by what it was asked for', async () => {
		const io = setup(row('real-id', 'Title', 'the-slug'))
		const connection = fakeConnection()

		renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['the-slug'] }),
		)

		await waitFor(() => expect(io.subscribe).toHaveBeenCalledTimes(1))
		expect(io.subscribe.mock.calls[0]?.[0]).toBe('real-id')
	})

	it('swaps in an update wholesale', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.change(row('a', 'renamed'))

		expect(result.current).toEqual({ status: Status.Ready, data: row('a', 'renamed') })
	})

	it('goes to gone on a delete, keeping only the title', async () => {
		const io = setup(row('a', 'Fix the nav'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.remove()

		expect(result.current).toEqual({ status: Status.Gone, title: 'Fix the nav' })
	})

	it('reports a load failure', async () => {
		const io = setup(row('a'))
		io.setServer(err(new Error('not found')))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)

		await waitFor(() => expect(result.current).toEqual({ status: Status.Failed, message: 'not found' }))
	})

	it('refetches once the subscription is open, closing the load race', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)

		await waitFor(() => expect(io.load).toHaveBeenCalledTimes(2))
	})

	it('refetches on reconnect', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(io.load).toHaveBeenCalledTimes(2))

		io.setServer(ok(row('a', 'changed while away')))
		act(() => connection.reconnect())

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, data: row('a', 'changed while away') }))
	})

	it('stays gone rather than being revived by a late update', async () => {
		const io = setup(row('a', 'Doomed'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.remove()
		await io.change(row('a', 'zombie'))

		expect(result.current).toEqual({ status: Status.Gone, title: 'Doomed' })
	})

	it('names the record by its latest title when it is deleted after a rename', async () => {
		const io = setup(row('a', 'Original'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.change(row('a', 'Renamed'))
		await io.remove()

		expect(result.current).toEqual({ status: Status.Gone, title: 'Renamed' })
	})

	it('stays idle and loads nothing when there is no record to load', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'], skip: true }),
		)

		expect(result.current).toEqual({ status: Status.Idle })
		expect(io.load).not.toHaveBeenCalled()
		expect(io.subscribe).not.toHaveBeenCalled()
	})

	it('closes the subscription on unmount', async () => {
		const io = setup(row('a'))
		const connection = fakeConnection()

		const { unmount } = renderHook(() =>
			useLiveRecord({ load: io.load, subscribe: io.subscribe, connection: connection.port, deps: ['a'] }),
		)
		await waitFor(() => expect(io.subscribe).toHaveBeenCalledTimes(1))

		unmount()

		await waitFor(() => expect(io.close).toHaveBeenCalledTimes(1))
	})
})

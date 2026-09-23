import { act, renderHook, waitFor } from '@testing-library/react';
import { foldUpdates } from '_/adapters/pocketbase/fold-updates';
import type { ConnectionPort } from '_/core/ports/connection';
import type { Unsubscribe } from '_/core/ports/subscription';
import { Status } from '_/lib/async-status';
import { ok } from '_/lib/result';
import type { ActionEvent } from '_/types';
import { describe, expect, it, vi } from 'vitest';
import { useLiveList } from './use-live-list';

type Row = { readonly id: string; readonly title: string }

const row = (id: string, title = id): Row => ({ id, title })

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

const setup = (rows: Row[]) => {
	let emit: (row: Row, action: ActionEvent) => void = () => {}
	const close = vi.fn(async () => {})
	let current = rows

	const list = vi.fn(async () => ok(current as readonly Row[]))
	const subscribe = vi.fn(async (update: (row: Row, action: ActionEvent) => void) => {
		emit = update
		return ok(close as Unsubscribe)
	})

	return {
		list,
		subscribe,
		close,
		emit: (row: Row, action: ActionEvent) => act(() => emit(row, action)),
		setServerRows: (next: Row[]) => {
			current = next
		},
	}
}

describe('useLiveList', () => {
	it('loads the list', async () => {
		const io = setup([row('a'), row('b')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, data: [row('a'), row('b')] }))
	})

	it('folds a create into the list', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.emit(row('b'), 'create')

		expect(result.current).toEqual({ status: Status.Ready, data: [row('a'), row('b')] })
	})

	it('folds an update in place', async () => {
		const io = setup([row('a'), row('b')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.emit(row('a', 'renamed'), 'update')

		expect(result.current).toEqual({ status: Status.Ready, data: [row('a', 'renamed'), row('b')] })
	})

	it('folds a delete out of the list', async () => {
		const io = setup([row('a'), row('b')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.emit(row('a'), 'delete')

		expect(result.current).toEqual({ status: Status.Ready, data: [row('b')] })
	})

	it('refetches after the subscription opens, closing the load race', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)

		await waitFor(() => expect(io.list).toHaveBeenCalledTimes(2))
	})

	it('refetches on reconnect, since nothing replays the gap', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(io.list).toHaveBeenCalledTimes(2))

		io.setServerRows([row('a'), row('c')])
		act(() => connection.reconnect())

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, data: [row('a'), row('c')] }))
	})

	it('ignores an event that arrives before the list is ready', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		await io.emit(row('z'), 'update')

		expect(result.current).toEqual({ status: Status.Ready, data: [row('a'), row('z')] })
	})

	it('neither loads nor subscribes while skipped', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() =>
			useLiveList({
				load: io.list,
				subscribe: io.subscribe,
				fold: foldUpdates,
				connection: connection.port,
				deps: [],
				skip: true,
			}),
		)
		await act(async () => {})

		expect(result.current).toEqual({ status: Status.Loading })
		expect(io.list).not.toHaveBeenCalled()
		expect(io.subscribe).not.toHaveBeenCalled()
	})

	it('loads and subscribes once it stops being skipped', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { result, rerender } = renderHook(
			({ skip }) =>
				useLiveList({
					load: io.list,
					subscribe: io.subscribe,
					fold: foldUpdates,
					connection: connection.port,
					deps: [],
					skip,
				}),
			{ initialProps: { skip: true } },
		)

		rerender({ skip: false })

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, data: [row('a')] }))
		expect(io.subscribe).toHaveBeenCalledTimes(1)
	})

	it('closes the subscription on unmount', async () => {
		const io = setup([row('a')])
		const connection = fakeConnection()

		const { unmount } = renderHook(() =>
			useLiveList({ load: io.list, subscribe: io.subscribe, fold: foldUpdates, connection: connection.port, deps: [] }),
		)
		await waitFor(() => expect(io.subscribe).toHaveBeenCalledTimes(1))

		unmount()

		await waitFor(() => expect(io.close).toHaveBeenCalledTimes(1))
	})
})

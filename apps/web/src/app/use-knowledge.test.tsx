import { act, renderHook, waitFor } from '@testing-library/react'
import type { ConnectionPort } from '_/core/ports/connection'
import type { Knowledge, KnowledgePort } from '_/core/ports/knowledge'
import type { Unsubscribe } from '_/core/ports/subscription'
import { Status } from '_/lib/async-status'
import { ok } from '_/lib/result'
import { describe, expect, it, vi } from 'vitest'
import { ContainerProvider, type Container } from './container'
import { useKnowledge, useKnowledgeList } from './use-knowledge'

const note = (id: string, title = id): Knowledge => ({
	id,
	slug: id,
	title,
	body: '',
	projectId: '',
	tags: [],
	createdAt: new Date(0),
	updatedAt: new Date(0),
})

const fakeConnection = () => {
	const listeners = new Set<() => void>()

	return {
		port: {
			onReconnect: (listener: () => void) => {
				listeners.add(listener)
				return () => listeners.delete(listener)
			},
			retry: () => {},
			onStatusChange: () => () => {},
		} satisfies ConnectionPort,
		reconnect: () => {
			for (const listener of listeners) listener()
		},
	}
}

const fakePort = (initial: Knowledge[]) => {
	let rows = initial
	let listChanged: () => void = () => {}
	let recordChanged: (note: Knowledge) => void = () => {}
	let recordGone: () => void = () => {}
	const close = vi.fn(async () => {})

	const port = {
		list: vi.fn(async () => ok(rows as readonly Knowledge[])),
		get: vi.fn(async (ref: string) => ok(rows.find(row => row.id === ref || row.slug === ref) ?? note(ref))),
		subscribeToList: vi.fn(async (onChange: () => void) => {
			listChanged = onChange
			return ok(close as Unsubscribe)
		}),
		subscribeToRecord: vi.fn(async (_id: string, onChange: (note: Knowledge) => void, onGone: () => void) => {
			recordChanged = onChange
			recordGone = onGone
			return ok(close as Unsubscribe)
		}),
	} satisfies KnowledgePort

	return {
		port,
		close,
		setServer: (next: Knowledge[]) => {
			rows = next
		},
		changeList: () => act(() => listChanged()),
		changeRecord: (next: Knowledge) => act(() => recordChanged(next)),
		remove: () => act(() => recordGone()),
	}
}

const wrapperFor = (knowledge: KnowledgePort, connection: ConnectionPort) => {
	const container = { knowledge, connection } as unknown as Container

	return ({ children }: { readonly children: React.ReactNode }) => (
		<ContainerProvider container={container}>{children}</ContainerProvider>
	)
}

describe('useKnowledgeList', () => {
	it('reloads the list when any note changes', async () => {
		const io = fakePort([note('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() => useKnowledgeList('term'), { wrapper: wrapperFor(io.port, connection.port) })
		await waitFor(() => expect(io.port.subscribeToList).toHaveBeenCalledTimes(1))
		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, notes: [note('a')] }))

		io.setServer([note('a'), note('b')])
		await io.changeList()

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, notes: [note('a'), note('b')] }))
		expect(io.port.list).toHaveBeenLastCalledWith({ text: 'term' })
	})

	it('reloads the list on reconnect', async () => {
		const io = fakePort([note('a')])
		const connection = fakeConnection()

		const { result } = renderHook(() => useKnowledgeList(''), { wrapper: wrapperFor(io.port, connection.port) })
		await waitFor(() => expect(result.current.status).toBe(Status.Ready))

		io.setServer([note('b')])
		act(() => connection.reconnect())

		await waitFor(() => expect(result.current).toEqual({ status: Status.Ready, notes: [note('b')] }))
	})

	it('closes the subscription on unmount', async () => {
		const io = fakePort([])
		const connection = fakeConnection()

		const { unmount } = renderHook(() => useKnowledgeList(''), { wrapper: wrapperFor(io.port, connection.port) })
		await waitFor(() => expect(io.port.subscribeToList).toHaveBeenCalledTimes(1))

		unmount()

		await waitFor(() => expect(io.close).toHaveBeenCalledTimes(1))
	})
})

describe('useKnowledge', () => {
	it('swaps in an update to the note', async () => {
		const io = fakePort([note('a', 'Old')])
		const connection = fakeConnection()

		const { result } = renderHook(() => useKnowledge('a'), { wrapper: wrapperFor(io.port, connection.port) })
		await waitFor(() => expect(io.port.subscribeToRecord).toHaveBeenCalledTimes(1))

		await io.changeRecord(note('a', 'New'))

		expect(result.current).toEqual({ status: Status.Ready, note: note('a', 'New') })
	})

	it('goes to gone when the note is deleted', async () => {
		const io = fakePort([note('a', 'Tip')])
		const connection = fakeConnection()

		const { result } = renderHook(() => useKnowledge('a'), { wrapper: wrapperFor(io.port, connection.port) })
		await waitFor(() => expect(io.port.subscribeToRecord).toHaveBeenCalledTimes(1))

		await io.remove()

		expect(result.current).toEqual({ status: Status.Gone, title: 'Tip' })
	})
})

import { act, renderHook, waitFor } from '@testing-library/react'
import { err, ok, type Result } from '_/lib/result'
import { describe, expect, it, vi } from 'vitest'
import { useAsyncState } from './use-async-state'
import { Status } from '_/lib/async-status'

const deferred = <T>() => {
	let resolve: (value: T) => void = () => {}
	const promise = new Promise<T>(settle => {
		resolve = settle
	})
	return { promise, resolve }
}

describe('useAsyncState', () => {
	it('starts loading and lands on the loaded value', async () => {
		const load = vi.fn(async () => ok('first'))

		const { result } = renderHook(() => useAsyncState(load, []))

		expect(result.current.state).toEqual({ status: Status.Loading })
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'first' }))
	})

	it('reports a failure with its message', async () => {
		const load = vi.fn(async () => err(new Error('offline')))

		const { result } = renderHook(() => useAsyncState(load, []))

		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Failed, message: 'offline' }))
	})

	it('refetches without flicking back to loading', async () => {
		let value = 'first'
		const load = vi.fn(async () => ok(value))

		const { result } = renderHook(() => useAsyncState(load, []))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'first' }))

		value = 'second'
		await act(async () => {
			await result.current.refetch()
		})

		expect(result.current.state).toEqual({ status: Status.Ready, data: 'second' })
		expect(load).toHaveBeenCalledTimes(2)
	})

	it('reloads through loading when a dependency changes', async () => {
		const load = vi.fn(async ({ id }: { id: string }) => ok(id))

		const { result, rerender } = renderHook(({ id }) => useAsyncState(() => load({ id }), [id]), {
			initialProps: { id: 'a' },
		})
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'a' }))

		rerender({ id: 'b' })

		expect(result.current.state).toEqual({ status: Status.Loading })
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'b' }))
	})

	it('ignores a load that resolves after the hook unmounted', async () => {
		const gate = deferred<Result<string>>()
		const load = vi.fn(() => gate.promise)

		const { unmount } = renderHook(() => useAsyncState(load, []))
		unmount()

		expect(() => gate.resolve(ok('late'))).not.toThrow()
	})

	it('keeps the newest result when an older load resolves last', async () => {
		const first = deferred<Result<string>>()
		const second = deferred<Result<string>>()
		const loads = [first.promise, second.promise]
		let call = 0
		const load = vi.fn(() => loads[call++] as Promise<Result<string>>)

		const { result, rerender } = renderHook(({ id }) => useAsyncState(load, [id]), {
			initialProps: { id: 'a' },
		})
		rerender({ id: 'b' })

		second.resolve(ok('newest'))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'newest' }))

		first.resolve(ok('stale'))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'newest' }))
	})

	it('patches the loaded value in place', async () => {
		const load = vi.fn(async () => ok(['a']))

		const { result } = renderHook(() => useAsyncState(load, []))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: ['a'] }))

		act(() => result.current.patch(rows => [...rows, 'b']))

		expect(result.current.state).toEqual({ status: Status.Ready, data: ['a', 'b'] })
	})

	it('ignores a patch while still loading, since there is nothing to patch', async () => {
		const gate = deferred<Result<string[]>>()
		const load = vi.fn(() => gate.promise)

		const { result } = renderHook(() => useAsyncState(load, []))

		act(() => result.current.patch(rows => [...rows, 'b']))

		expect(result.current.state).toEqual({ status: Status.Loading })

		gate.resolve(ok(['a']))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: ['a'] }))
	})

	it('does not load while skipped', async () => {
		const load = vi.fn(async () => ok('value'))

		const { result } = renderHook(() => useAsyncState(load, [], true))

		expect(load).not.toHaveBeenCalled()
		expect(result.current.state).toEqual({ status: Status.Loading })
	})

	it('loads once it stops being skipped', async () => {
		const load = vi.fn(async () => ok('value'))

		const { result, rerender } = renderHook(({ skip }) => useAsyncState(load, [], skip), {
			initialProps: { skip: true },
		})
		expect(load).not.toHaveBeenCalled()

		rerender({ skip: false })

		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'value' }))
	})

	it('hands back a stable refetch across renders', async () => {
		const load = vi.fn(async () => ok('value'))

		const { result, rerender } = renderHook(() => useAsyncState(load, []))
		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: 'value' }))

		const before = result.current.refetch
		rerender()

		expect(result.current.refetch).toBe(before)
	})
})

import { renderHook, waitFor } from '@testing-library/react'
import { err, ok, type Result } from '_/lib/result'
import type { Unsubscribe } from '_/core/ports/subscription'
import { describe, expect, it, vi } from 'vitest'
import { useSubscription } from './use-subscription'

const deferred = <T>() => {
	let resolve: (value: T) => void = () => {}
	const promise = new Promise<T>(settle => {
		resolve = settle
	})
	return { promise, resolve }
}

describe('useSubscription', () => {
	it('opens the subscription once mounted', async () => {
		const close = vi.fn(async () => {})
		const open = vi.fn(async () => ok(close as Unsubscribe))

		renderHook(() => useSubscription(open, []))

		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))
	})

	it('closes the subscription on unmount', async () => {
		const close = vi.fn(async () => {})
		const open = vi.fn(async () => ok(close as Unsubscribe))

		const { unmount } = renderHook(() => useSubscription(open, []))
		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))

		unmount()

		await waitFor(() => expect(close).toHaveBeenCalledTimes(1))
	})

	it('closes a subscription that opened after the hook was already gone', async () => {
		const close = vi.fn(async () => {})
		const gate = deferred<Result<Unsubscribe>>()
		const open = vi.fn(() => gate.promise)

		const { unmount } = renderHook(() => useSubscription(open, []))
		unmount()
		gate.resolve(ok(close as Unsubscribe))

		await waitFor(() => expect(close).toHaveBeenCalledTimes(1))
	})

	it('swallows a failure to open rather than throwing into the tree', async () => {
		const open = vi.fn(async () => err(new Error('offline')))

		expect(() => renderHook(() => useSubscription(open, []))).not.toThrow()

		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))
	})

	it('reopens when a dependency changes, closing the previous one', async () => {
		const close = vi.fn(async () => {})
		const open = vi.fn(async () => ok(close as Unsubscribe))

		const { rerender } = renderHook(({ id }) => useSubscription(open, [id]), {
			initialProps: { id: 'a' },
		})
		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))

		rerender({ id: 'b' })

		await waitFor(() => expect(open).toHaveBeenCalledTimes(2))
		await waitFor(() => expect(close).toHaveBeenCalledTimes(1))
	})

	it('does not reopen when the dependencies are unchanged', async () => {
		const close = vi.fn(async () => {})
		const open = vi.fn(async () => ok(close as Unsubscribe))

		const { rerender } = renderHook(({ id }) => useSubscription(open, [id]), {
			initialProps: { id: 'a' },
		})
		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))

		rerender({ id: 'a' })

		expect(open).toHaveBeenCalledTimes(1)
	})

	it('survives a close that rejects', async () => {
		const close = vi.fn(async () => {
			throw new Error('already gone')
		})
		const open = vi.fn(async () => ok(close as Unsubscribe))

		const { unmount } = renderHook(() => useSubscription(open, []))
		await waitFor(() => expect(open).toHaveBeenCalledTimes(1))

		expect(() => unmount()).not.toThrow()
		await waitFor(() => expect(close).toHaveBeenCalledTimes(1))
	})
})

import { act, renderHook } from '@testing-library/react'
import type { SharedItem } from '_/core/domain/share'
import type { SharesPort } from '_/core/ports/shares'
import { Status } from '_/lib/async-status'
import { err, ok, type Result } from '_/lib/result'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ContainerProvider, type Container } from './container'
import { SHARE_POLL_MS, useShared } from './use-shared'

const plan = (title: string): SharedItem => ({
	kind: 'plan',
	label: 'vendor',
	plan: { title, goal: '', status: 'active', tags: [], progress: { total: 0, done: 0 }, updatedAt: new Date(0) },
	todos: [],
})

type Answer = Result<SharedItem | undefined>

const render = (open: SharesPort['open']) => {
	const container = { shares: { open } } as unknown as Container
	const wrapper = ({ children }: { readonly children: React.ReactNode }) => (
		<ContainerProvider container={container}>{children}</ContainerProvider>
	)

	return renderHook(() => useShared('tok'), { wrapper })
}

const settle = () => act(async () => {})

const tick = (ms = SHARE_POLL_MS) =>
	act(async () => {
		await vi.advanceTimersByTimeAsync(ms)
	})

const setVisibility = (state: DocumentVisibilityState) => {
	Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => state })
	document.dispatchEvent(new Event('visibilitychange'))
}

describe('useShared', () => {
	beforeEach(() => {
		vi.useFakeTimers()
		setVisibility('visible')
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('opens the shared item', async () => {
		const open = vi.fn(async (): Promise<Answer> => ok(plan('Migrate auth')))

		const { result } = render(open)
		expect(result.current.status).toBe(Status.Loading)
		await settle()

		expect(result.current).toEqual({ status: Status.Ready, item: plan('Migrate auth') })
		expect(open).toHaveBeenCalledWith('tok')
	})

	it('picks up a change on the next poll', async () => {
		let title = 'Migrate auth'
		const open = vi.fn(async (): Promise<Answer> => ok(plan(title)))

		const { result } = render(open)
		await settle()
		title = 'Migrate auth, again'

		await tick(SHARE_POLL_MS - 1)
		expect(open).toHaveBeenCalledTimes(1)

		await tick(1)
		expect(open).toHaveBeenCalledTimes(2)
		expect(result.current).toEqual({ status: Status.Ready, item: plan('Migrate auth, again') })
	})

	it('shows a revoked link as gone and stops polling', async () => {
		const answers: Answer[] = [ok(plan('Migrate auth')), ok(undefined)]
		const open = vi.fn(async (): Promise<Answer> => answers.shift() ?? ok(plan('resurrected')))

		const { result } = render(open)
		await settle()
		await tick()

		expect(result.current).toEqual({ status: Status.Gone })

		await tick()
		await tick()
		expect(open).toHaveBeenCalledTimes(2)
		expect(result.current).toEqual({ status: Status.Gone })
	})

	it('keeps the last snapshot when a poll fails', async () => {
		const answers: Answer[] = [ok(plan('Migrate auth')), err(new Error('offline'))]
		const open = vi.fn(async (): Promise<Answer> => answers.shift() ?? ok(plan('back online')))

		const { result } = render(open)
		await settle()
		await tick()

		expect(result.current).toEqual({ status: Status.Ready, item: plan('Migrate auth') })

		await tick()
		expect(result.current).toEqual({ status: Status.Ready, item: plan('back online') })
	})

	it('fails when the first load fails', async () => {
		const open = vi.fn(async (): Promise<Answer> => err(new Error('Cannot reach the server.')))

		const { result } = render(open)
		await settle()

		expect(result.current).toEqual({ status: Status.Failed, message: 'Cannot reach the server.' })
	})

	it('does not poll a hidden tab, and catches up when it is shown', async () => {
		const open = vi.fn(async (): Promise<Answer> => ok(plan('Migrate auth')))

		render(open)
		await settle()
		setVisibility('hidden')

		await tick()
		await tick()
		expect(open).toHaveBeenCalledTimes(1)

		await act(async () => setVisibility('visible'))
		expect(open).toHaveBeenCalledTimes(2)
	})

	it('ignores an answer that arrives after a newer one', async () => {
		const pending: Array<(answer: Answer) => void> = []
		const open = vi.fn(() => new Promise<Answer>(resolve => pending.push(resolve)))

		const { result } = render(open)
		await tick()
		expect(pending).toHaveLength(2)

		await act(async () => pending[1]?.(ok(plan('newer'))))
		await act(async () => pending[0]?.(ok(plan('older'))))

		expect(result.current).toEqual({ status: Status.Ready, item: plan('newer') })
	})

	it('stops polling once unmounted', async () => {
		const open = vi.fn(async (): Promise<Answer> => ok(plan('Migrate auth')))

		const { unmount } = render(open)
		await settle()
		unmount()

		await tick()
		setVisibility('visible')
		expect(open).toHaveBeenCalledTimes(1)
	})
})

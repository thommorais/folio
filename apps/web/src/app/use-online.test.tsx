import { act, renderHook } from '@testing-library/react'
import type { ConnectionPort } from '_/core/ports/connection'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ContainerProvider, type Container } from './container'
import { useOnline } from './use-online'

const setup = () => {
	let emit: (online: boolean) => void = () => {}
	const connection: ConnectionPort = {
		onReconnect: () => () => {},
		retry: () => {},
		onStatusChange: listener => {
			emit = listener
			return () => {}
		},
	}
	const container = { connection } as unknown as Container
	const wrapper = ({ children }: { readonly children: React.ReactNode }) => (
		<ContainerProvider container={container}>{children}</ContainerProvider>
	)
	const { result } = renderHook(() => useOnline(), { wrapper })

	return { result, emit: (online: boolean) => act(() => emit(online)) }
}

describe('useOnline', () => {
	beforeEach(() => vi.useFakeTimers())
	afterEach(() => vi.useRealTimers())

	it('starts online', () => {
		expect(setup().result.current).toBe(true)
	})

	it('ignores a drop that recovers within the grace period', () => {
		const io = setup()

		io.emit(false)
		act(() => {
			vi.advanceTimersByTime(1000)
		})
		io.emit(true)
		act(() => {
			vi.advanceTimersByTime(2000)
		})

		expect(io.result.current).toBe(true)
	})

	it('reports offline once the drop outlasts the grace period', () => {
		const io = setup()

		io.emit(false)
		act(() => {
			vi.advanceTimersByTime(2000)
		})

		expect(io.result.current).toBe(false)
	})

	it('reports online again as soon as the connection is back', () => {
		const io = setup()
		io.emit(false)
		act(() => {
			vi.advanceTimersByTime(2000)
		})

		io.emit(true)

		expect(io.result.current).toBe(true)
	})
})

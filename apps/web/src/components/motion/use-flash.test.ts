import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useFlash } from './use-flash'

describe('useFlash', () => {
	beforeEach(() => vi.useFakeTimers())
	afterEach(() => vi.useRealTimers())

	it('stays quiet when the row first appears', () => {
		const { result } = renderHook(({ value }) => useFlash(value), { initialProps: { value: 1 } })

		expect(result.current).toBe(false)
	})

	it('lights up when the watched value changes', () => {
		const { result, rerender } = renderHook(({ value }) => useFlash(value), { initialProps: { value: 1 } })

		rerender({ value: 2 })

		expect(result.current).toBe(true)
	})

	it('fades back after a moment', () => {
		const { result, rerender } = renderHook(({ value }) => useFlash(value), { initialProps: { value: 1 } })

		rerender({ value: 2 })
		act(() => {
			vi.advanceTimersByTime(1200)
		})

		expect(result.current).toBe(false)
	})

	it('ignores a re-render with the same value', () => {
		const { result, rerender } = renderHook(({ value }) => useFlash(value), { initialProps: { value: 1 } })

		rerender({ value: 1 })

		expect(result.current).toBe(false)
	})
})

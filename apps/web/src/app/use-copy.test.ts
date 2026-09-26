import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useCopy } from './use-copy'

describe('useCopy', () => {
	const writeText = vi.fn(async () => {})

	beforeEach(() => {
		vi.useFakeTimers()
		Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
	})

	afterEach(() => {
		vi.useRealTimers()
		writeText.mockClear()
	})

	it('writes the text and reports it copied', async () => {
		const { result } = renderHook(() => useCopy())

		await act(async () => {
			await result.current.copy('folio todo create')
		})

		expect(writeText).toHaveBeenCalledWith('folio todo create')
		expect(result.current.copied).toBe(true)
	})

	it('clears the copied state after two seconds', async () => {
		const { result } = renderHook(() => useCopy())

		await act(async () => {
			await result.current.copy('x')
		})
		act(() => {
			vi.advanceTimersByTime(2000)
		})

		expect(result.current.copied).toBe(false)
	})
})

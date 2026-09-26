import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { capturedKey, useTypeToSearch } from './type-to-search'

const outside = { closest: () => null }
const inside = { closest: () => ({}) }

const key = (value: string, over: Partial<Parameters<typeof capturedKey>[0]> = {}) => ({
	key: value,
	metaKey: false,
	ctrlKey: false,
	altKey: false,
	target: outside,
	...over,
})

describe('capturedKey', () => {
	it('takes a printable character typed outside a field', () => {
		expect(capturedKey(key('a'))).toEqual({ insert: 'a' })
	})

	it('focuses without inserting on slash', () => {
		expect(capturedKey(key('/'))).toEqual({ insert: '' })
	})

	it('leaves shortcuts with a modifier alone', () => {
		expect(capturedKey(key('k', { metaKey: true }))).toBeUndefined()
		expect(capturedKey(key('c', { ctrlKey: true }))).toBeUndefined()
		expect(capturedKey(key('a', { altKey: true }))).toBeUndefined()
	})

	it('leaves keys typed into a field, menu or dialog alone', () => {
		expect(capturedKey(key('a', { target: inside }))).toBeUndefined()
	})

	it('ignores non-printing keys and space', () => {
		for (const value of ['Enter', 'Escape', 'ArrowDown', 'Tab', ' ']) {
			expect(capturedKey(key(value))).toBeUndefined()
		}
	})
})

describe('useTypeToSearch', () => {
	const setup = () => {
		const input = document.createElement('input')
		document.body.append(input)
		const onType = vi.fn()
		const hook = renderHook(() => useTypeToSearch({ current: input }, onType))

		return {
			input,
			onType,
			unmount: () => {
				hook.unmount()
				input.remove()
			},
		}
	}

	it('focuses the search and hands over the typed character', () => {
		const { input, onType, unmount } = setup()

		document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'a', bubbles: true }))

		expect(document.activeElement).toBe(input)
		expect(onType).toHaveBeenCalledWith('a')
		unmount()
	})

	it('lets a real modifier shortcut through', () => {
		const { input, onType, unmount } = setup()

		document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true }))

		expect(document.activeElement).not.toBe(input)
		expect(onType).not.toHaveBeenCalled()
		unmount()
	})
})

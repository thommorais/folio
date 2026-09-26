import { afterEach, describe, expect, it } from 'vitest'
import { renderHook } from '@testing-library/react'
import { nextRow, useRowNavigation } from './row-navigation'

describe('nextRow', () => {
	it('starts at the first row going down and the last going up', () => {
		expect(nextRow(-1, 3, 'ArrowDown')).toBe(0)
		expect(nextRow(-1, 3, 'ArrowUp')).toBe(2)
	})

	it('moves one row at a time', () => {
		expect(nextRow(0, 3, 'ArrowDown')).toBe(1)
		expect(nextRow(2, 3, 'ArrowUp')).toBe(1)
	})

	it('stops at the ends instead of wrapping', () => {
		expect(nextRow(2, 3, 'ArrowDown')).toBe(2)
		expect(nextRow(0, 3, 'ArrowUp')).toBe(0)
	})

	it('has nowhere to go in an empty list', () => {
		expect(nextRow(-1, 0, 'ArrowDown')).toBeUndefined()
	})
})

describe('useRowNavigation', () => {
	afterEach(() => {
		document.body.innerHTML = ''
	})

	const rows = () => {
		document.body.innerHTML = `
			<a href="/a" data-row>a</a>
			<a href="/b" data-row>b</a>
			<input id="field" />
		`
		return [...document.querySelectorAll<HTMLElement>('[data-row]')]
	}

	const press = (key: string, target: EventTarget = document.body) =>
		target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true }))

	it('moves focus down the rows with the arrow keys', () => {
		const [first, second] = rows()
		renderHook(() => useRowNavigation())

		press('ArrowDown')
		expect(document.activeElement).toBe(first)

		press('ArrowDown', first)
		expect(document.activeElement).toBe(second)
	})

	it('leaves arrows typed into a field alone', () => {
		rows()
		const field = document.getElementById('field') as HTMLInputElement
		field.focus()
		renderHook(() => useRowNavigation())

		press('ArrowDown', field)

		expect(document.activeElement).toBe(field)
	})
})

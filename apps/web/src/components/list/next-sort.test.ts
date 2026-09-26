import { describe, expect, it } from 'vitest'
import { nextSort } from './next-sort'

describe('nextSort', () => {
	it.each([
		['title', 'asc'],
		['slug', 'asc'],
		['status', 'asc'],
		['position', 'asc'],
		['priority', 'desc'],
		['size', 'desc'],
		['created', 'desc'],
		['updated', 'desc'],
	] as const)('starts %s %s', (field, direction) => {
		expect(nextSort(undefined, field)).toEqual({ field, direction })
	})

	it('flips the direction of the active field', () => {
		expect(nextSort({ field: 'title', direction: 'asc' }, 'title')).toEqual({ field: 'title', direction: 'desc' })
		expect(nextSort({ field: 'created', direction: 'desc' }, 'created')).toEqual({ field: 'created', direction: 'asc' })
	})

	it('switches to another field in that field’s own starting direction', () => {
		expect(nextSort({ field: 'created', direction: 'desc' }, 'title')).toEqual({ field: 'title', direction: 'asc' })
	})
})

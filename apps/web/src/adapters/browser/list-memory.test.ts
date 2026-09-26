import { beforeEach, describe, expect, it } from 'vitest'
import { createListMemory } from './list-memory'

describe('listMemory', () => {
	beforeEach(() => localStorage.clear())

	it('restores the filters last used on a list', () => {
		const memory = createListMemory()
		memory.remember('tickets', { statuses: ['done'], sort: { field: 'title', direction: 'asc' } })

		expect(memory.restore('tickets', {})).toEqual({ statuses: ['done'], sort: { field: 'title', direction: 'asc' } })
	})

	it('keeps each list separate', () => {
		const memory = createListMemory()
		memory.remember('tickets', { tags: ['web'] })

		expect(memory.restore('todos', {})).toBeUndefined()
	})

	it('leaves a link that already carries filters alone', () => {
		const memory = createListMemory()
		memory.remember('tickets', { tags: ['web'] })

		expect(memory.restore('tickets', { priority: 'high' })).toBeUndefined()
	})

	it('remembers clearing the filters, so they do not come back', () => {
		const memory = createListMemory()
		memory.remember('tickets', { tags: ['web'] })
		memory.remember('tickets', {})

		expect(memory.restore('tickets', {})).toBeUndefined()
	})

	it('forgets the search text and record ids, which belong to one visit or one project', () => {
		const memory = createListMemory()
		memory.remember('todos', { q: 'nav', ticket: 'abc', plan: 'def', tags: ['web'] })

		expect(memory.restore('todos', {})).toEqual({ tags: ['web'] })
	})

	it('ignores a corrupt stored value', () => {
		localStorage.setItem('folio.list-filters', 'not json')

		expect(createListMemory().restore('tickets', {})).toBeUndefined()
	})
})

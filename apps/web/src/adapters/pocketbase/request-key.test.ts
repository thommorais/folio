import { describe, expect, it } from 'vitest'
import { keyed } from './request-key'

const keyOf = (scope: string, options: Record<string, unknown>, distinguish?: unknown): string =>
	keyed(scope, options, distinguish).requestKey as string

describe('keyed', () => {
	it('separates two queries against the same collection', () => {
		const tickets = keyOf('issues.count', { filter: 'kind = "ticket"' })
		const todos = keyOf('issues.count', { filter: 'kind = "todo"' })

		expect(tickets).not.toBe(todos)
	})

	it('gives the same query the same key, so a repeat still cancels the one in flight', () => {
		const first = keyOf('issues.list', { filter: 'kind = "todo"', sort: '-created' })
		const second = keyOf('issues.list', { filter: 'kind = "todo"', sort: '-created' })

		expect(first).toBe(second)
	})

	it('ignores the order the options were written in', () => {
		const a = keyOf('issues.list', { filter: 'x', sort: 'title', expand: 'project' })
		const b = keyOf('issues.list', { expand: 'project', sort: 'title', filter: 'x' })

		expect(a).toBe(b)
	})

	it('separates the same query made by different callers', () => {
		const count = keyOf('issues.count', { filter: 'kind = "ticket"' })
		const list = keyOf('issues.list', { filter: 'kind = "ticket"' })

		expect(count).not.toBe(list)
	})

	it('treats an absent option and an undefined one as the same query', () => {
		const absent = keyOf('issues.list', { filter: 'x' })
		const undef = keyOf('issues.list', { filter: 'x', sort: undefined })

		expect(absent).toBe(undef)
	})

	it('distinguishes nested values rather than collapsing them', () => {
		const one = keyOf('issues.list', { params: { kinds: ['journal', 'doc'] } })
		const two = keyOf('issues.list', { params: { kinds: ['doc', 'journal'] } })

		expect(one).not.toBe(two)
	})

	it('keeps the options it was given', () => {
		const options = keyed('issues.count', { filter: 'kind = "ticket"' })

		expect(options.filter).toBe('kind = "ticket"')
	})

	it('separates two reads that differ only outside the query', () => {
		const page = keyOf('issues.list', { filter: 'x' }, [50, 0])
		const next = keyOf('issues.list', { filter: 'x' }, [50, 50])

		expect(page).not.toBe(next)
	})

	// parentId narrows the rows after they arrive, so the two reads issue the
	// same query and would otherwise cancel each other.
	it('separates reads that differ only by a filter applied after the response', () => {
		const all = keyOf('issues.list', { filter: 'kind = "ticket"' }, [undefined, undefined, undefined])
		const children = keyOf('issues.list', { filter: 'kind = "ticket"' }, [undefined, undefined, 'map-id'])

		expect(all).not.toBe(children)
	})

	it('leaves the distinguishing values out of the options sent to the API', () => {
		const options = keyed('issues.list', { filter: 'x' }, [50, 100])

		expect(Object.keys(options).sort()).toEqual(['filter', 'requestKey'])
	})

	it('keys the same read the same way when nothing distinguishes it', () => {
		expect(keyOf('issues.list', { filter: 'x' })).toBe(keyOf('issues.list', { filter: 'x' }, undefined))
	})
})

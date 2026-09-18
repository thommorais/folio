import { describe, expect, it } from 'vitest'
import { keyed } from './request-key'

const keyOf = (scope: string, options: Record<string, unknown>): string => keyed(scope, options).requestKey as string

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
})

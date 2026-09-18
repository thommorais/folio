import { describe, expect, it } from 'vitest'
import { partitionChildren } from './frontier'
import type { Issue, IssueId } from './issue'

const issue = (id: string, over: Partial<Issue> = {}): Issue =>
	({
		id: id as IssueId,
		projectId: 'p1',
		parentId: 'map' as IssueId,
		slug: id,
		title: id,
		body: '',
		status: 'open',
		priority: 'medium',
		assignee: undefined,
		tags: [],
		externalRef: '',
		dependsOn: [],
		wayfinder: undefined,
		createdBy: undefined,
		createdAt: new Date('2026-01-01'),
		updatedAt: new Date('2026-01-01'),
		...over,
	}) as Issue

describe('partitionChildren', () => {
	it('puts an open, unblocked, unassigned child on the frontier', () => {
		const got = partitionChildren([issue('a')])

		expect(got.frontier.map(t => t.id)).toEqual(['a'])
		expect(got.blocked).toHaveLength(0)
	})

	it('holds a child back while a blocker is open', () => {
		const got = partitionChildren([issue('a'), issue('b', { dependsOn: ['a' as IssueId], blocked: true })])

		expect(got.frontier.map(t => t.id)).toEqual(['a'])
		expect(got.blocked.map(t => t.id)).toEqual(['b'])
	})

	it('releases a child once every blocker is terminal', () => {
		const got = partitionChildren([
			issue('a', { status: 'done' }),
			issue('b', { dependsOn: ['a' as IssueId], blocked: false }),
		])

		expect(got.frontier.map(t => t.id)).toEqual(['b'])
		expect(got.done.map(t => t.id)).toEqual(['a'])
	})

	it('separates a claimed child from the frontier', () => {
		const got = partitionChildren([issue('a', { assignee: 'u1' as Issue['assignee'] })])

		expect(got.frontier).toHaveLength(0)
		expect(got.claimed.map(t => t.id)).toEqual(['a'])
	})

	it('ignores a blocker that is not among the children', () => {
		const got = partitionChildren([issue('b', { dependsOn: ['ghost' as IssueId] })])

		expect(got.frontier.map(t => t.id)).toEqual(['b'])
	})

	it('orders the frontier oldest first', () => {
		const got = partitionChildren([
			issue('late', { createdAt: new Date('2026-03-01') }),
			issue('early', { createdAt: new Date('2026-01-01') }),
		])

		expect(got.frontier.map(t => t.id)).toEqual(['early', 'late'])
	})
})

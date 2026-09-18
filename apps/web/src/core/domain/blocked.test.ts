import { describe, expect, it } from 'vitest'
import { withBlocked } from './blocked'
import type { Issue, IssueId } from './issue'

const issue = (id: string, over: Partial<Issue> = {}): Issue =>
	({
		id: id as IssueId,
		projectId: 'p1',
		parentId: undefined,
		slug: id,
		title: id,
		body: '',
		status: 'open',
		priority: 'medium',
		assignee: undefined,
		tags: [],
		externalRef: '',
		dependsOn: [],
		relatedTo: [],
		blocked: false,
		wayfinder: undefined,
		createdBy: undefined,
		createdAt: new Date('2026-01-01'),
		updatedAt: new Date('2026-01-01'),
		...over,
	}) as Issue

const blockedOf = (issues: readonly Issue[]) => new Map(withBlocked(issues).map(one => [one.id as string, one.blocked]))

describe('withBlocked', () => {
	it('blocks an issue whose blocker is still open', () => {
		const got = blockedOf([issue('a'), issue('b', { dependsOn: ['a' as IssueId] })])

		expect(got.get('b')).toBe(true)
	})

	it('releases an issue once its blocker is done', () => {
		const got = blockedOf([issue('a', { status: 'done' }), issue('b', { dependsOn: ['a' as IssueId] })])

		expect(got.get('b')).toBe(false)
	})

	it('treats a cancelled blocker as cleared', () => {
		const got = blockedOf([issue('a', { status: 'cancelled' }), issue('b', { dependsOn: ['a' as IssueId] })])

		expect(got.get('b')).toBe(false)
	})

	it('holds an issue back while any one of its blockers is open', () => {
		const got = blockedOf([
			issue('a', { status: 'done' }),
			issue('b'),
			issue('c', { dependsOn: ['a' as IssueId, 'b' as IssueId] }),
		])

		expect(got.get('c')).toBe(true)
	})

	// The adapter cannot judge a blocker it did not load, so an unknown one
	// does not block. Recomputing has to keep that rule or a filtered list
	// would show everything as takeable.
	it('ignores a blocker that is not in the set', () => {
		const got = blockedOf([issue('b', { dependsOn: ['ghost' as IssueId] })])

		expect(got.get('b')).toBe(false)
	})

	it('corrects a stale flag rather than trusting it', () => {
		const got = blockedOf([issue('a', { status: 'done' }), issue('b', { dependsOn: ['a' as IssueId], blocked: true })])

		expect(got.get('b')).toBe(false)
	})

	it('leaves an issue with no blockers alone', () => {
		expect(blockedOf([issue('a')]).get('a')).toBe(false)
	})

	it('returns the same issue objects when nothing changed', () => {
		const issues = [issue('a'), issue('b')]

		expect(withBlocked(issues)[0]).toBe(issues[0])
	})
})

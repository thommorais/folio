import { describe, expect, it } from 'vitest'
import { layout, subtreeOf } from './graph'
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

const child = (id: string, parent: string, over: Partial<Issue> = {}) =>
	issue(id, { parentId: parent as IssueId, ...over })

const ids = (issues: readonly Issue[]) => issues.map(one => one.id).sort()

describe('subtreeOf', () => {
	it('returns the root alone when it has no children', () => {
		expect(ids(subtreeOf([issue('root'), issue('other')], 'root' as IssueId))).toEqual(['root'])
	})

	it('collects descendants at every depth', () => {
		const pool = [issue('root'), child('a', 'root'), child('b', 'a'), child('c', 'b'), issue('elsewhere')]

		expect(ids(subtreeOf(pool, 'root' as IssueId))).toEqual(['a', 'b', 'c', 'root'])
	})

	it('leaves out a sibling branch', () => {
		const pool = [issue('root'), child('a', 'root'), issue('other'), child('x', 'other')]

		expect(ids(subtreeOf(pool, 'root' as IssueId))).toEqual(['a', 'root'])
	})

	it('returns nothing when the root is not in the pool', () => {
		expect(subtreeOf([issue('a')], 'ghost' as IssueId)).toEqual([])
	})

	it('survives a parent cycle without looping forever', () => {
		const pool = [
			issue('root'),
			child('a', 'root'),
			child('b', 'a'),
			// b is a's parent as well as its child: a cycle the writer should have
			// refused, kept survivable here so the view still renders.
			issue('a2', { parentId: 'b' as IssueId }),
		]

		expect(ids(subtreeOf(pool, 'root' as IssueId))).toEqual(['a', 'a2', 'b', 'root'])
	})
})

describe('layout', () => {
	it('places a lone node at rank zero', () => {
		const { nodes } = layout([issue('a')])

		expect(nodes.map(node => [node.issue.id, node.rank])).toEqual([['a', 0]])
	})

	it('ranks a child below its parent', () => {
		const { nodes } = layout([issue('root'), child('a', 'root')])
		const rank = new Map(nodes.map(node => [node.issue.id, node.rank]))

		expect(rank.get('root' as IssueId)).toBe(0)
		expect(rank.get('a' as IssueId)).toBe(1)
	})

	it('ranks a blocked issue below its blocker', () => {
		const { nodes } = layout([issue('a'), issue('b', { dependsOn: ['a' as IssueId] })])
		const rank = new Map(nodes.map(node => [node.issue.id, node.rank]))

		expect(rank.get('b' as IssueId)).toBeGreaterThan(rank.get('a' as IssueId) ?? 0)
	})

	it('takes the longest path when parent and blocker disagree', () => {
		// c's parent sits at rank 0, but its blocker b is at rank 1, so c must
		// land at 2 rather than 1.
		const { nodes } = layout([
			issue('root'),
			child('a', 'root'),
			issue('b', { dependsOn: ['a' as IssueId] }),
			child('c', 'root', { dependsOn: ['b' as IssueId] }),
		])
		const rank = new Map(nodes.map(node => [node.issue.id, node.rank]))

		expect(rank.get('c' as IssueId)).toBe(3)
	})

	it('gives every node on a rank a distinct order', () => {
		const { nodes } = layout([issue('root'), child('a', 'root'), child('b', 'root'), child('c', 'root')])
		const orders = nodes.filter(node => node.rank === 1).map(node => node.order)

		expect(new Set(orders).size).toBe(3)
	})

	it('emits one parent edge per child', () => {
		const { edges } = layout([issue('root'), child('a', 'root')])

		expect(edges).toEqual([{ from: 'root', to: 'a', kind: 'parent' }])
	})

	it('emits a blocks edge from the blocker to the blocked', () => {
		const { edges } = layout([issue('a'), issue('b', { dependsOn: ['a' as IssueId] })])

		expect(edges).toEqual([{ from: 'a', to: 'b', kind: 'blocks' }])
	})

	it('emits one relates edge for a reciprocated relation', () => {
		const { edges } = layout([issue('a', { relatedTo: ['b' as IssueId] }), issue('b', { relatedTo: ['a' as IssueId] })])

		expect(edges).toEqual([{ from: 'a', to: 'b', kind: 'relates' }])
	})

	it('drops an edge whose other end is not in the set', () => {
		const { edges } = layout([issue('a', { dependsOn: ['ghost' as IssueId], relatedTo: ['ghost' as IssueId] })])

		expect(edges).toEqual([])
	})

	it('ranks a dependency cycle rather than looping forever', () => {
		const { nodes } = layout([issue('a', { dependsOn: ['b' as IssueId] }), issue('b', { dependsOn: ['a' as IssueId] })])

		expect(nodes).toHaveLength(2)
		expect(nodes.every(node => Number.isFinite(node.rank))).toBe(true)
	})
})

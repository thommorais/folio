import { describe, expect, it } from 'vitest'
import { NODE, place } from './graph-geometry'
import type { Graph, PlacedNode } from '_/core/domain/graph'
import type { Issue, IssueId } from '_/core/domain/issue'

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

const node = (id: string, rank: number, order: number): PlacedNode => ({ issue: issue(id), rank, order })

const graph = (nodes: readonly PlacedNode[], edges: Graph['edges'] = []): Graph => ({ nodes, edges })

describe('place', () => {
	it('puts a lone node at the origin of the canvas', () => {
		const { boxes } = place(graph([node('a', 0, 0)]))

		expect(boxes[0]?.x).toBe(0)
		expect(boxes[0]?.y).toBe(0)
	})

	it('drops each rank below the one above it', () => {
		const { boxes } = place(graph([node('a', 0, 0), node('b', 1, 0)]))
		const y = new Map(boxes.map(box => [box.id, box.y]))

		expect(y.get('b' as IssueId)).toBeGreaterThan(y.get('a' as IssueId) ?? 0)
	})

	it('spreads nodes on one rank across, not down', () => {
		const { boxes } = place(graph([node('a', 0, 0), node('b', 0, 1)]))

		expect(boxes[0]?.y).toBe(boxes[1]?.y)
		expect(boxes[1]?.x).toBeGreaterThan(boxes[0]?.x ?? 0)
	})

	it('centres a short rank against a wider one', () => {
		// One node above three: it should sit over the middle of them, not hug
		// the left edge.
		const { boxes, width } = place(graph([node('top', 0, 0), node('a', 1, 0), node('b', 1, 1), node('c', 1, 2)]))
		const top = boxes.find(box => box.id === 'top')

		expect((top?.x ?? 0) + NODE.width / 2).toBeCloseTo(width / 2)
	})

	it('sizes the canvas to hold every node', () => {
		const { boxes, width, height } = place(graph([node('a', 0, 0), node('b', 1, 1)]))

		for (const box of boxes) {
			expect(box.x + NODE.width).toBeLessThanOrEqual(width)
			expect(box.y + NODE.height).toBeLessThanOrEqual(height)
		}
	})

	it('routes an edge from the bottom of one node to the top of the other', () => {
		const { lines } = place(
			graph([node('a', 0, 0), node('b', 1, 0)], [{ from: 'a' as IssueId, to: 'b' as IssueId, kind: 'parent' }]),
		)

		expect(lines).toHaveLength(1)
		expect(lines[0]?.from.y).toBeLessThan(lines[0]?.to.y ?? 0)
	})

	it('drops an edge whose end was never placed', () => {
		const { lines } = place(
			graph([node('a', 0, 0)], [{ from: 'a' as IssueId, to: 'ghost' as IssueId, kind: 'blocks' }]),
		)

		expect(lines).toEqual([])
	})

	it('keeps the edge kind so it can be drawn differently', () => {
		const { lines } = place(
			graph([node('a', 0, 0), node('b', 1, 0)], [{ from: 'a' as IssueId, to: 'b' as IssueId, kind: 'blocks' }]),
		)

		expect(lines[0]?.kind).toBe('blocks')
	})

	it('has nothing to place for an empty graph', () => {
		const { boxes, lines, width, height } = place(graph([]))

		expect(boxes).toEqual([])
		expect(lines).toEqual([])
		expect(width).toBe(0)
		expect(height).toBe(0)
	})
})

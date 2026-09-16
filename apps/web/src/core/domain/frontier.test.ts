import { describe, expect, it } from 'vitest'
import { partitionChildren } from './frontier'
import type { Ticket, TicketId } from './ticket'

const ticket = (id: string, over: Partial<Ticket> = {}): Ticket =>
	({
		id: id as TicketId,
		projectId: 'p1',
		parentId: 'map' as TicketId,
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
	}) as Ticket

describe('partitionChildren', () => {
	it('puts an open, unblocked, unassigned child on the frontier', () => {
		const got = partitionChildren([ticket('a')])

		expect(got.frontier.map(t => t.id)).toEqual(['a'])
		expect(got.blocked).toHaveLength(0)
	})

	it('holds a child back while a blocker is open', () => {
		const got = partitionChildren([ticket('a'), ticket('b', { dependsOn: ['a' as TicketId] })])

		expect(got.frontier.map(t => t.id)).toEqual(['a'])
		expect(got.blocked.map(t => t.id)).toEqual(['b'])
	})

	it('releases a child once every blocker is terminal', () => {
		const got = partitionChildren([
			ticket('a', { status: 'closed' }),
			ticket('b', { dependsOn: ['a' as TicketId] }),
		])

		expect(got.frontier.map(t => t.id)).toEqual(['b'])
		expect(got.done.map(t => t.id)).toEqual(['a'])
	})

	it('separates a claimed child from the frontier', () => {
		const got = partitionChildren([ticket('a', { assignee: 'u1' as Ticket['assignee'] })])

		expect(got.frontier).toHaveLength(0)
		expect(got.claimed.map(t => t.id)).toEqual(['a'])
	})

	it('ignores a blocker that is not among the children', () => {
		const got = partitionChildren([ticket('b', { dependsOn: ['ghost' as TicketId] })])

		expect(got.frontier.map(t => t.id)).toEqual(['b'])
	})

	it('orders the frontier oldest first', () => {
		const got = partitionChildren([
			ticket('late', { createdAt: new Date('2026-03-01') }),
			ticket('early', { createdAt: new Date('2026-01-01') }),
		])

		expect(got.frontier.map(t => t.id)).toEqual(['early', 'late'])
	})
})

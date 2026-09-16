import { describe, expect, it } from 'vitest'
import { buildTicketTree } from './ticket-tree'
import type { Ticket, TicketId } from './ticket'

const ticket = (id: string, over: Partial<Ticket> = {}): Ticket =>
	({
		id: id as TicketId,
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
		wayfinder: undefined,
		createdBy: undefined,
		createdAt: new Date('2026-01-01'),
		updatedAt: new Date('2026-01-01'),
		...over,
	}) as Ticket

const child = (id: string, parent: string, over: Partial<Ticket> = {}) =>
	ticket(id, { parentId: parent as TicketId, ...over })

describe('buildTicketTree', () => {
	it('keeps a flat list flat, at depth zero', () => {
		const got = buildTicketTree([ticket('a'), ticket('b')])

		expect(got.map(row => [row.ticket.id, row.depth])).toEqual([
			['a', 0],
			['b', 0],
		])
	})

	it('nests a child under its parent', () => {
		const got = buildTicketTree([ticket('root'), child('kid', 'root')])

		expect(got.map(row => [row.ticket.id, row.depth])).toEqual([
			['root', 0],
			['kid', 1],
		])
	})

	it('nests to arbitrary depth', () => {
		const got = buildTicketTree([ticket('a'), child('b', 'a'), child('c', 'b'), child('d', 'c')])

		expect(got.map(row => row.depth)).toEqual([0, 1, 2, 3])
	})

	it('emits children directly after their parent, not after later roots', () => {
		const got = buildTicketTree([ticket('a'), ticket('z'), child('a1', 'a')])

		expect(got.map(row => row.ticket.id)).toEqual(['a', 'a1', 'z'])
	})

	// A child whose parent is filtered out would otherwise vanish from the list
	// entirely, so it is promoted rather than dropped.
	it('promotes a child to a root when its parent is absent', () => {
		const got = buildTicketTree([child('orphan', 'missing')])

		expect(got.map(row => [row.ticket.id, row.depth])).toEqual([['orphan', 0]])
	})

	it('marks the last child of a parent so the connector can close the run', () => {
		const got = buildTicketTree([ticket('a'), child('b', 'a'), child('c', 'a')])

		expect(got.map(row => [row.ticket.id, row.isLast])).toEqual([
			['a', true],
			['b', false],
			['c', true],
		])
	})

	it('survives a parent cycle without hanging', () => {
		const got = buildTicketTree([child('a', 'b'), child('b', 'a')])

		expect(got).toHaveLength(2)
	})

	it('preserves the order the list arrived in, so the active sort still applies', () => {
		const got = buildTicketTree([ticket('z'), ticket('a'), child('z2', 'z'), child('z1', 'z')])

		expect(got.map(row => row.ticket.id)).toEqual(['z', 'z2', 'z1', 'a'])
	})
})

describe('buildTicketTree with context ancestors', () => {
	// A search matching only a deep child still shows where that child lives.
	it('pulls in a non-matching parent as context', () => {
		const got = buildTicketTree([child('kid', 'root')], { context: [ticket('root')] })

		expect(got.map(row => [row.ticket.id, row.depth, row.isContext])).toEqual([
			['root', 0, true],
			['kid', 1, false],
		])
	})

	it('pulls in a whole ancestor chain', () => {
		const got = buildTicketTree([child('c', 'b')], {
			context: [ticket('a'), child('b', 'a')],
		})

		expect(got.map(row => [row.ticket.id, row.isContext])).toEqual([
			['a', true],
			['b', true],
			['c', false],
		])
	})

	it('does not mark a ticket as context when it is itself a match', () => {
		const got = buildTicketTree([ticket('root'), child('kid', 'root')], { context: [ticket('root')] })

		expect(got.map(row => [row.ticket.id, row.isContext])).toEqual([
			['root', false],
			['kid', false],
		])
	})

	it('ignores context that no match depends on', () => {
		const got = buildTicketTree([ticket('a')], { context: [ticket('unrelated')] })

		expect(got.map(row => row.ticket.id)).toEqual(['a'])
	})
})

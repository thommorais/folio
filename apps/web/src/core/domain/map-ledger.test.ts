import { describe, expect, it } from 'vitest'
import type { Cycle, CycleId } from './cycle'
import type { Issue, IssueId } from './issue'
import { mapLedger, planHold } from './map-ledger'

const issue = (id: string, over: Partial<Issue> = {}): Issue =>
	({
		id: id as IssueId,
		kind: 'ticket',
		projectId: 'p1',
		parentId: 'map' as IssueId,
		slug: id,
		title: id,
		status: 'open',
		wayfinder: 'grilling',
		resolution: '',
		resolutionEntry: undefined,
		dependsOn: [],
		createdAt: new Date('2026-01-01'),
		updatedAt: new Date('2026-01-01'),
		...over,
	}) as Issue

const cycle = (over: Partial<Cycle> = {}): Cycle =>
	({
		id: 'c1' as CycleId,
		ordinal: 1,
		phase: 'plan',
		resolution: '',
		mapId: 'map' as IssueId,
		closedAt: undefined,
		...over,
	}) as Cycle

describe('mapLedger', () => {
	it('lists done decisions as decided and cancelled ones as out of scope, oldest first', () => {
		const got = mapLedger([
			issue('late', { status: 'done', resolution: 'B', createdAt: new Date('2026-02-01') }),
			issue('open'),
			issue('early', { status: 'done', resolution: 'A' }),
			issue('dropped', { status: 'cancelled', resolution: 'Past the destination' }),
		])

		expect(got.decided.map(d => [d.id, d.resolution])).toEqual([
			['early', 'A'],
			['late', 'B'],
		])
		expect(got.outOfScope.map(d => d.id)).toEqual(['dropped'])
	})

	it('leaves todos out: they are steps, not decisions', () => {
		const got = mapLedger([issue('step', { kind: 'todo', status: 'done' })])

		expect(got.decided).toHaveLength(0)
	})
})

describe('planHold', () => {
	it('counts the open decisions holding a mapped cycle in plan', () => {
		const children = [issue('a'), issue('b', { status: 'in_progress' }), issue('c', { status: 'done' })]

		expect(planHold(cycle(), children)).toBe(2)
	})

	it('holds nothing once the cycle has left plan, is resolved, or has no map', () => {
		const children = [issue('a')]

		expect(planHold(cycle({ phase: 'do' }), children)).toBe(0)
		expect(planHold(cycle({ closedAt: new Date() }), children)).toBe(0)
		expect(planHold(cycle({ mapId: undefined }), children)).toBe(0)
	})
})

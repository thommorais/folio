import { describe, expect, it } from 'vitest'
import { ISSUE_KIND, type Issue, type IssueId } from './issue'
import type { Plan, PlanId } from './plan'
import { groupPlans, groupTodos } from './grouping'

const ticket = (id: string): Issue => ({ id: id as IssueId, kind: ISSUE_KIND.TICKET, title: id }) as Issue

const todo = (id: string, over: Partial<Issue> = {}): Issue =>
	({ id: id as IssueId, kind: ISSUE_KIND.TODO, title: id, planId: undefined, parentId: undefined, ...over }) as Issue

const plan = (id: string, over: Partial<Plan> = {}): Plan =>
	({ id: id as PlanId, title: id, ticketId: undefined, ...over }) as Plan

const keys = (groups: ReturnType<typeof groupTodos>) =>
	groups.map(group => [group.target?.kind ?? 'none', group.target?.id ?? '', group.items.map(item => item.id)])

describe('groupTodos', () => {
	it('groups a todo under its plan', () => {
		const groups = groupTodos([todo('t1', { planId: 'p1' as PlanId })], [plan('p1')], [])

		expect(keys(groups)).toEqual([['plan', 'p1', ['t1']]])
	})

	it('groups a todo with no plan under its ticket parent', () => {
		const groups = groupTodos([todo('t1', { parentId: 'k1' as IssueId })], [], [ticket('k1')])

		expect(keys(groups)).toEqual([['ticket', 'k1', ['t1']]])
	})

	it('prefers the plan when a todo has both', () => {
		const groups = groupTodos(
			[todo('t1', { planId: 'p1' as PlanId, parentId: 'k1' as IssueId })],
			[plan('p1')],
			[ticket('k1')],
		)

		expect(keys(groups)).toEqual([['plan', 'p1', ['t1']]])
	})

	it('puts a todo with neither in a trailing ungrouped block', () => {
		const groups = groupTodos([todo('t0'), todo('t1', { parentId: 'k1' as IssueId })], [], [ticket('k1')])

		expect(keys(groups)).toEqual([
			['ticket', 'k1', ['t1']],
			['none', '', ['t0']],
		])
	})

	it('treats a parent that cannot be resolved as ungrouped', () => {
		const groups = groupTodos([todo('t1', { planId: 'gone' as PlanId })], [], [])

		expect(keys(groups)).toEqual([['none', '', ['t1']]])
	})

	it('keeps groups in order of first appearance and items in input order', () => {
		const groups = groupTodos(
			[todo('t1', { parentId: 'k2' as IssueId }), todo('t2', { parentId: 'k1' as IssueId }), todo('t3', { parentId: 'k2' as IssueId })],
			[],
			[ticket('k1'), ticket('k2')],
		)

		expect(keys(groups)).toEqual([
			['ticket', 'k2', ['t1', 't3']],
			['ticket', 'k1', ['t2']],
		])
	})
})

describe('groupPlans', () => {
	it('groups plans by ticket and leaves ticketless plans last', () => {
		const groups = groupPlans(
			[plan('p0'), plan('p1', { ticketId: 'k1' as IssueId }), plan('p2', { ticketId: 'k1' as IssueId })],
			[ticket('k1')],
		)

		expect(keys(groups)).toEqual([
			['ticket', 'k1', ['p1', 'p2']],
			['none', '', ['p0']],
		])
	})

	it('treats an unresolved ticket as ungrouped', () => {
		const groups = groupPlans([plan('p1', { ticketId: 'gone' as IssueId })], [])

		expect(keys(groups)).toEqual([['none', '', ['p1']]])
	})
})

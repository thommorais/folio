import { anEntry, anIssue } from '_/test/records'
import { describe, expect, it } from 'vitest'
import type { Plan, PlanId } from './plan'
import type { ProjectId } from './project'
import { sortEntries, sortIssues, sortPlans } from './order'

const ids = (rows: readonly { readonly id: string }[]) => rows.map(row => row.id)

const day = (n: number) => new Date(2026, 0, n)

const aPlan = (id: string, over: Partial<Plan> = {}): Plan => ({
	id: id as PlanId,
	projectId: 'p1' as ProjectId,
	ticketId: undefined,
	title: id,
	goal: '',
	status: 'active',
	tags: [],
	createdBy: undefined,
	createdAt: day(1),
	updatedAt: day(1),
	...over,
})

describe('sortIssues', () => {
	it('puts high priority first when sorting by priority descending', () => {
		const rows = [
			anIssue('m', { priority: 'medium' }),
			anIssue('h', { priority: 'high' }),
			anIssue('l', { priority: 'low' }),
		]

		expect(ids(sortIssues(rows, { field: 'priority', direction: 'desc' }))).toEqual(['h', 'm', 'l'])
	})

	it('orders status by what needs attention, not alphabetically', () => {
		const rows = [
			anIssue('done', { status: 'done' }),
			anIssue('open', { status: 'open' }),
			anIssue('cancelled', { status: 'cancelled' }),
			anIssue('blocked', { status: 'blocked' }),
			anIssue('doing', { status: 'in_progress' }),
		]

		expect(ids(sortIssues(rows, { field: 'status', direction: 'asc' }))).toEqual([
			'doing',
			'open',
			'blocked',
			'done',
			'cancelled',
		])
	})

	it('falls back to newest first when no sort is given', () => {
		const rows = [anIssue('old', { createdAt: day(1) }), anIssue('new', { createdAt: day(3) }), anIssue('mid', { createdAt: day(2) })]

		expect(ids(sortIssues(rows, undefined))).toEqual(['new', 'mid', 'old'])
	})

	it('breaks ties newest first', () => {
		const rows = [
			anIssue('old', { priority: 'high', createdAt: day(1) }),
			anIssue('new', { priority: 'high', createdAt: day(2) }),
		]

		expect(ids(sortIssues(rows, { field: 'priority', direction: 'desc' }))).toEqual(['new', 'old'])
	})

	it('sorts titles by locale, ignoring case', () => {
		const rows = [anIssue('b', { title: 'beta' }), anIssue('a', { title: 'Alpha' }), anIssue('c', { title: 'charlie' })]

		expect(ids(sortIssues(rows, { field: 'title', direction: 'asc' }))).toEqual(['a', 'b', 'c'])
	})

	it('keeps unsized issues last in either direction', () => {
		const rows = [anIssue('none'), anIssue('small', { size: 1 }), anIssue('big', { size: 8 })]

		expect(ids(sortIssues(rows, { field: 'size', direction: 'asc' }))).toEqual(['small', 'big', 'none'])
		expect(ids(sortIssues(rows, { field: 'size', direction: 'desc' }))).toEqual(['big', 'small', 'none'])
	})

	it('groups tickets by status, then priority, when no sort is chosen', () => {
		const rows = [
			anIssue('open-low', { kind: 'ticket', status: 'open', priority: 'low' }),
			anIssue('doing-medium', { kind: 'ticket', status: 'in_progress', priority: 'medium' }),
			anIssue('open-high', { kind: 'ticket', status: 'open', priority: 'high' }),
			anIssue('blocked-high', { kind: 'ticket', status: 'blocked', priority: 'high' }),
			anIssue('doing-high', { kind: 'ticket', status: 'in_progress', priority: 'high' }),
		]

		expect(ids(sortIssues(rows, undefined, 'ticket'))).toEqual([
			'doing-high',
			'doing-medium',
			'open-high',
			'open-low',
			'blocked-high',
		])
	})

	it('keeps todos in their manual position when no sort is chosen', () => {
		const rows = [
			anIssue('third', { kind: 'todo', position: 3, createdAt: day(3) }),
			anIssue('first', { kind: 'todo', position: 1, createdAt: day(1) }),
			anIssue('second', { kind: 'todo', position: 2, createdAt: day(2) }),
		]

		expect(ids(sortIssues(rows, undefined, 'todo'))).toEqual(['first', 'second', 'third'])
	})

	it('lets a chosen sort override the default order', () => {
		const rows = [anIssue('a', { title: 'b', position: 1 }), anIssue('b', { title: 'a', position: 2 })]

		expect(ids(sortIssues(rows, { field: 'title', direction: 'asc' }, 'todo'))).toEqual(['b', 'a'])
	})

	it('does not mutate its input', () => {
		const rows = [anIssue('a', { position: 2 }), anIssue('b', { position: 1 })]

		sortIssues(rows, { field: 'position', direction: 'asc' })

		expect(ids(rows)).toEqual(['a', 'b'])
	})
})

describe('sortPlans', () => {
	it('ranks active before draft before closed plans', () => {
		const rows = [
			aPlan('abandoned', { status: 'abandoned' }),
			aPlan('draft', { status: 'draft' }),
			aPlan('done', { status: 'done' }),
			aPlan('active', { status: 'active' }),
		]

		expect(ids(sortPlans(rows, { field: 'status', direction: 'asc' }))).toEqual(['active', 'draft', 'done', 'abandoned'])
	})
})

describe('sortEntries', () => {
	it('sorts by last update', () => {
		const rows = [anEntry('a', { updatedAt: day(1) }), anEntry('b', { updatedAt: day(5) }), anEntry('c', { updatedAt: day(3) })]

		expect(ids(sortEntries(rows, { field: 'updated', direction: 'desc' }))).toEqual(['b', 'c', 'a'])
	})
})

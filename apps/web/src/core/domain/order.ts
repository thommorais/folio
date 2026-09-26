import { SORT_DIRECTION, type EntrySortField, type IssueSortField, type PlanSortField, type Sort } from '_/core/ports/sort'
import type { Entry } from './entry'
import { ISSUE_STATUS, PRIORITY, type Issue, type IssueStatus, type Priority } from './issue'
import { PLAN_STATUS, type Plan, type PlanStatus } from './plan'

type Key = number | string | undefined

type Row = { readonly id: string; readonly createdAt: Date }

const PRIORITY_RANK: Record<Priority, number> = { [PRIORITY.LOW]: 1, [PRIORITY.MEDIUM]: 2, [PRIORITY.HIGH]: 3 }

const ISSUE_STATUS_RANK: Record<IssueStatus, number> = {
	[ISSUE_STATUS.IN_PROGRESS]: 0,
	[ISSUE_STATUS.OPEN]: 1,
	[ISSUE_STATUS.BLOCKED]: 2,
	[ISSUE_STATUS.DONE]: 3,
	[ISSUE_STATUS.CANCELLED]: 4,
}

const PLAN_STATUS_RANK: Record<PlanStatus, number> = {
	[PLAN_STATUS.ACTIVE]: 0,
	[PLAN_STATUS.DRAFT]: 1,
	[PLAN_STATUS.DONE]: 2,
	[PLAN_STATUS.ABANDONED]: 3,
}

const NEWEST_FIRST = { field: 'created', direction: SORT_DIRECTION.DESC } as const

const compareKeys = (a: Key, b: Key): number => {
	if (typeof a === 'string' && typeof b === 'string') return a.localeCompare(b, undefined, { sensitivity: 'base' })
	return (a as number) - (b as number)
}

const sortBy = <T extends Row, F extends string>(
	rows: readonly T[],
	sort: Sort<F> | undefined,
	keys: Record<F, (row: T) => Key>,
	fallback: Sort<F>,
): readonly T[] => {
	const { field, direction } = sort ?? fallback
	const keyOf = keys[field]
	const sign = direction === SORT_DIRECTION.DESC ? -1 : 1

	return [...rows].sort((a, b) => {
		const left = keyOf(a)
		const right = keyOf(b)

		if (left === undefined || right === undefined) {
			if (left !== right) return left === undefined ? 1 : -1
		} else {
			const order = compareKeys(left, right) * sign
			if (order !== 0) return order
		}

		return b.createdAt.getTime() - a.createdAt.getTime() || a.id.localeCompare(b.id)
	})
}

const time = (date: Date) => date.getTime()

const ISSUE_KEYS: Record<IssueSortField, (issue: Issue) => Key> = {
	position: issue => issue.position,
	title: issue => issue.title,
	status: issue => ISSUE_STATUS_RANK[issue.status],
	priority: issue => PRIORITY_RANK[issue.priority],
	size: issue => issue.size,
	created: issue => time(issue.createdAt),
	updated: issue => time(issue.updatedAt),
}

const PLAN_KEYS: Record<PlanSortField, (plan: Plan) => Key> = {
	title: plan => plan.title,
	status: plan => PLAN_STATUS_RANK[plan.status],
	created: plan => time(plan.createdAt),
	updated: plan => time(plan.updatedAt),
}

const ENTRY_KEYS: Record<EntrySortField, (entry: Entry) => Key> = {
	title: entry => entry.title,
	slug: entry => entry.slug,
	created: entry => time(entry.createdAt),
	updated: entry => time(entry.updatedAt),
}

export const sortIssues = (rows: readonly Issue[], sort: Sort<IssueSortField> | undefined) =>
	sortBy(rows, sort, ISSUE_KEYS, NEWEST_FIRST)

export const sortPlans = (rows: readonly Plan[], sort: Sort<PlanSortField> | undefined) =>
	sortBy(rows, sort, PLAN_KEYS, NEWEST_FIRST)

export const sortEntries = (rows: readonly Entry[], sort: Sort<EntrySortField> | undefined) =>
	sortBy(rows, sort, ENTRY_KEYS, NEWEST_FIRST)

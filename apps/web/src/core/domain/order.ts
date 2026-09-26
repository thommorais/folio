import { SORT_DIRECTION, type EntrySortField, type IssueSortField, type PlanSortField, type Sort } from '_/core/ports/sort'
import type { Entry } from './entry'
import { ISSUE_KIND, ISSUE_STATUS, PRIORITY, type Issue, type IssueKind, type IssueStatus, type Priority } from './issue'
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

const NEWEST_FIRST = [{ field: 'created', direction: SORT_DIRECTION.DESC }] as const

export const DEFAULT_ISSUE_ORDER: Record<IssueKind, readonly Sort<IssueSortField>[]> = {
	[ISSUE_KIND.TICKET]: [
		{ field: 'status', direction: SORT_DIRECTION.ASC },
		{ field: 'priority', direction: SORT_DIRECTION.DESC },
	],
	[ISSUE_KIND.TODO]: [{ field: 'position', direction: SORT_DIRECTION.ASC }],
}

const compareKeys = (a: Key, b: Key): number => {
	if (typeof a === 'string' && typeof b === 'string') return a.localeCompare(b, undefined, { sensitivity: 'base' })
	return (a as number) - (b as number)
}

const compareBy =
	<T extends Row, F extends string>(keys: Record<F, (row: T) => Key>, { field, direction }: Sort<F>) =>
	(a: T, b: T): number => {
		const left = keys[field](a)
		const right = keys[field](b)

		if (left === undefined || right === undefined) return left === right ? 0 : left === undefined ? 1 : -1

		return compareKeys(left, right) * (direction === SORT_DIRECTION.DESC ? -1 : 1)
	}

const sortBy = <T extends Row, F extends string>(
	rows: readonly T[],
	order: readonly Sort<F>[],
	keys: Record<F, (row: T) => Key>,
): readonly T[] => {
	const comparators = order.map(sort => compareBy(keys, sort))

	return [...rows].sort((a, b) => {
		for (const compare of comparators) {
			const result = compare(a, b)
			if (result !== 0) return result
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

export const sortIssues = (rows: readonly Issue[], sort: Sort<IssueSortField> | undefined, kind?: IssueKind) =>
	sortBy(rows, sort ? [sort] : kind ? DEFAULT_ISSUE_ORDER[kind] : NEWEST_FIRST, ISSUE_KEYS)

export const sortPlans = (rows: readonly Plan[], sort: Sort<PlanSortField> | undefined) =>
	sortBy(rows, sort ? [sort] : NEWEST_FIRST, PLAN_KEYS)

export const sortEntries = (rows: readonly Entry[], sort: Sort<EntrySortField> | undefined) =>
	sortBy(rows, sort ? [sort] : NEWEST_FIRST, ENTRY_KEYS)

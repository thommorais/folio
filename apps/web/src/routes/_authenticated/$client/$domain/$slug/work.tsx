import { createFileRoute } from '@tanstack/react-router'
import { ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue'
import { PLAN_STATUSES, type PlanStatus } from '_/core/domain/plan'
import { WORK_SORT_FIELDS, type Sort, type WorkSortField } from '_/core/ports/sort'
import { WORK_TYPES, type WorkType } from '_/pages/work/types'
import { Work } from '_/pages/work'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type WorkSearch = {
	readonly q?: string
	readonly types?: readonly WorkType[]
	// Issues and plans name their statuses differently, and both use "done", so
	// each half is kept in its own param rather than one ambiguous list.
	readonly statuses?: readonly IssueStatus[]
	readonly planStatuses?: readonly PlanStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly ticket?: string
	readonly sort?: Sort<WorkSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/work')({
	validateSearch: (search: Record<string, unknown>): WorkSearch => ({
		q: asString(search.q),
		types: asMembers(WORK_TYPES, search.types),
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		planStatuses: asMembers(PLAN_STATUSES, search.planStatuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		ticket: asString(search.ticket),
		sort: asSort(WORK_SORT_FIELDS, search.sort),
	}),
	component: Work,
})

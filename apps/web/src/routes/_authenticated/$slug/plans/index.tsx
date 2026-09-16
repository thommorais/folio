import { createFileRoute } from '@tanstack/react-router'
import { PLAN_STATUSES, type PlanStatus } from '_/core/domain/plan'
import { PLAN_SORT_FIELDS, type PlanSortField, type Sort } from '_/core/ports/sort'
import { Plans } from '_/pages/plans'
import { asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type PlansSearch = {
	readonly ticket?: string
	readonly statuses?: readonly PlanStatus[]
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<PlanSortField>
}

export const Route = createFileRoute('/_authenticated/$slug/plans/')({
	validateSearch: (search: Record<string, unknown>): PlansSearch => ({
		ticket: asString(search.ticket),
		statuses: asMembers(PLAN_STATUSES, search.statuses),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(PLAN_SORT_FIELDS, search.sort),
	}),
	component: Plans,
})

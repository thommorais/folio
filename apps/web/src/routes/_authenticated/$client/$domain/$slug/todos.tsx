import { createFileRoute } from '@tanstack/react-router'
import { ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue'
import { ISSUE_SORT_FIELDS, type IssueSortField, type Sort } from '_/core/ports/sort'
import { Issues } from '_/pages/issues'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type TodosSearch = {
	readonly todo?: string
	readonly ticket?: string
	readonly plan?: string
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<IssueSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/todos')({
	validateSearch: (search: Record<string, unknown>): TodosSearch => ({
		todo: asString(search.todo),
		ticket: asString(search.ticket),
		plan: asString(search.plan),
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(ISSUE_SORT_FIELDS, search.sort),
	}),
	component: () => <Issues kind='todo' emptyLabel='todos' />,
})

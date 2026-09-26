import { createFileRoute } from '@tanstack/react-router';
import { DEFAULT_ISSUE_STATUSES, ISSUE_KIND, ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue';
import { TODO_SORT_FIELDS, type Sort, type TodoSortField } from '_/core/ports/sort';
import { Issues } from '_/pages/issues';
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params';

export type TodosSearch = {
	readonly ticket?: string
	readonly plan?: string
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<TodoSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/todos/')({
	validateSearch: (search: Record<string, unknown>): TodosSearch => ({
		ticket: asString(search.ticket),
		plan: asString(search.plan),
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(TODO_SORT_FIELDS, search.sort),
	}),
	component: () => <Issues kind={ISSUE_KIND.TODO} emptyLabel='todos' defaultStatuses={DEFAULT_ISSUE_STATUSES} />,
})

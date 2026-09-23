import { createFileRoute } from '@tanstack/react-router'
import { DEFAULT_TICKET_STATUSES, ISSUE_STATUSES, PRIORITIES, ISSUE_KIND, type IssueStatus, type Priority } from '_/core/domain/issue'
import { ISSUE_SORT_FIELDS, type IssueSortField, type Sort } from '_/core/ports/sort'
import { Issues } from '_/pages/issues'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type TicketsSearch = {
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<IssueSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/tickets/')({
	validateSearch: (search: Record<string, unknown>): TicketsSearch => ({
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(ISSUE_SORT_FIELDS, search.sort),
	}),
	component: () => <Issues kind={ISSUE_KIND.TICKET} emptyLabel='tickets' defaultStatuses={DEFAULT_TICKET_STATUSES} />,
})

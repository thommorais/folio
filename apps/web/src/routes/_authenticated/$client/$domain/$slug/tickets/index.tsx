import { createFileRoute, redirect } from '@tanstack/react-router'
import { listMemory } from '_/adapters/browser/list-memory'
import { DEFAULT_ISSUE_STATUSES, ISSUE_STATUSES, PRIORITIES, ISSUE_KIND, type IssueStatus, type Priority } from '_/core/domain/issue'
import { TICKET_SORT_FIELDS, type Sort, type TicketSortField } from '_/core/ports/sort'
import { Issues } from '_/pages/issues'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type TicketsSearch = {
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<TicketSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/tickets/')({
	beforeLoad: ({ search, cause, params }) => {
		if (cause !== 'enter') return
		const restored = listMemory.restore('/_authenticated/$client/$domain/$slug/tickets/', search)
		if (restored) throw redirect({ to: '/$client/$domain/$slug/tickets', params, search: restored })
	},
	validateSearch: (search: Record<string, unknown>): TicketsSearch => ({
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(TICKET_SORT_FIELDS, search.sort),
	}),
	component: () => <Issues kind={ISSUE_KIND.TICKET} emptyLabel='tickets' defaultStatuses={DEFAULT_ISSUE_STATUSES} />,
})

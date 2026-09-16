import { createFileRoute } from '@tanstack/react-router'
import { PRIORITIES, type Priority } from '_/core/domain/todo'
import { TICKET_STATUSES, type TicketStatus } from '_/core/domain/ticket'
import { TICKET_SORT_FIELDS, type Sort, type TicketSortField } from '_/core/ports/sort'
import { Tickets } from '_/pages/tickets'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type TicketsSearch = {
	readonly statuses?: readonly TicketStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<TicketSortField>
}

export const Route = createFileRoute('/_authenticated/$slug/tickets/')({
	validateSearch: (search: Record<string, unknown>): TicketsSearch => ({
		statuses: asMembers(TICKET_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(TICKET_SORT_FIELDS, search.sort),
	}),
	component: Tickets,
})

import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Ticket } from '_/core/domain/ticket'
import type { TicketFilter } from '_/core/ports/tickets'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type TicketsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly tickets: readonly Ticket[] }
	| { readonly status: 'failed'; readonly message: string }

export const useTickets = (project: string, filter?: TicketFilter): TicketsState => {
	const { tickets, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Ticket>({
		load: () => tickets.list(project, JSON.parse(key) as TicketFilter),
		subscribe: update => tickets.subscribeToList(project, update, JSON.parse(key) as TicketFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, tickets],
	})

	return state.status === 'ready' ? { status: 'ready', tickets: state.data } : state
}

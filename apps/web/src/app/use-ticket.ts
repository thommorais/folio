import type { Ticket } from '_/core/domain/ticket'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type TicketState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly ticket: Ticket }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useTicket = (project: string, slug: string): TicketState => {
	const { tickets, connection } = useContainer()

	const state = useLiveRecord<Ticket>({
		load: () => tickets.get(project, slug),
		subscribe: (id, onChange, onGone) => tickets.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === 'ready' ? { status: 'ready', ticket: state.data } : state
}

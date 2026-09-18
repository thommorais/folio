import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Client } from '_/core/domain/client'
import type { ClientFilter } from '_/core/ports/clients'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type ClientsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly clients: readonly Client[] }
	| { readonly status: 'failed'; readonly message: string }

export const useClients = (filter?: ClientFilter): ClientsState => {
	const { clients, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Client>({
		load: () => clients.list(JSON.parse(key) as ClientFilter),
		subscribe: update => clients.subscribeToList(update, JSON.parse(key) as ClientFilter),
		fold: foldUpdates,
		connection,
		deps: [key, clients],
	})

	return state.status === 'ready' ? { status: 'ready', clients: state.data } : state
}

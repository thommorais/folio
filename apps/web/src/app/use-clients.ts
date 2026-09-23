import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Client } from '_/core/domain/client'
import type { ClientFilter } from '_/core/ports/clients'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type ClientsState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly clients: readonly Client[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

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

	return state.status === Status.Ready ? { status: Status.Ready, clients: state.data } : state
}

import type { Client } from '_/core/domain/client'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type ClientState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly client: Client }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useClient = (slug: string | undefined): ClientState => {
	const { clients, connection } = useContainer()

	const state = useLiveRecord<Client>({
		load: () => clients.get(slug ?? ''),
		subscribe: (id, onChange, onGone) => clients.subscribeToRecord(id, onChange, onGone),
		titleOf: client => client.name,
		connection,
		deps: [slug, clients],
		skip: !slug,
	})

	return state.status === Status.Ready ? { status: Status.Ready, client: state.data } : state
}

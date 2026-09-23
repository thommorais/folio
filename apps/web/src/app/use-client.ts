import type { Client } from '_/core/domain/client'
import { useAsyncState } from './realtime/use-async-state'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type ClientState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly client: Client }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useClient = (slug: string | undefined): ClientState => {
	const { clients } = useContainer()

	const { state } = useAsyncState<Client>(() => clients.get(slug ?? ''), [slug, clients], !slug)

	return state.status === Status.Ready ? { status: Status.Ready, client: state.data } : state
}

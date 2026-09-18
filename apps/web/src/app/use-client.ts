import type { Client } from '_/core/domain/client'
import { useAsyncState } from './realtime/use-async-state'
import { useContainer } from './container'

type ClientState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly client: Client }
	| { readonly status: 'failed'; readonly message: string }

export const useClient = (slug: string | undefined): ClientState => {
	const { clients } = useContainer()

	const { state } = useAsyncState<Client>(() => clients.get(slug ?? ''), [slug, clients], !slug)

	return state.status === 'ready' ? { status: 'ready', client: state.data } : state
}

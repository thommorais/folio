import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Domain } from '_/core/domain/domain'
import type { DomainFilter } from '_/core/ports/domains'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type DomainsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly domains: readonly Domain[] }
	| { readonly status: 'failed'; readonly message: string }

export const useDomains = (filter?: DomainFilter): DomainsState => {
	const { domains, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Domain>({
		load: () => domains.list(JSON.parse(key) as DomainFilter),
		subscribe: update => domains.subscribeToList(update, JSON.parse(key) as DomainFilter),
		fold: foldUpdates,
		connection,
		deps: [key, domains],
	})

	return state.status === 'ready' ? { status: 'ready', domains: state.data } : state
}

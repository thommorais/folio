import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Domain } from '_/core/domain/domain'
import type { DomainFilter } from '_/core/ports/domains'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type DomainsState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly domains: readonly Domain[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useDomains = (filter?: DomainFilter, skip = false): DomainsState => {
	const { domains, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Domain>({
		load: () => domains.list(JSON.parse(key) as DomainFilter),
		subscribe: update => domains.subscribeToList(update, JSON.parse(key) as DomainFilter),
		fold: foldUpdates,
		connection,
		deps: [key, domains],
		skip,
	})

	return state.status === Status.Ready ? { status: Status.Ready, domains: state.data } : state
}

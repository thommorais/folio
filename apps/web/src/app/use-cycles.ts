import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Cycle } from '_/core/domain/cycle'
import type { CycleFilter } from '_/core/ports/cycles'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type CyclesState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly cycles: readonly Cycle[] }
	| { readonly status: 'failed'; readonly message: string }

export const useCycles = (project: string, filter?: CycleFilter): CyclesState => {
	const { cycles, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Cycle>({
		load: () => cycles.list(project, JSON.parse(key) as CycleFilter),
		subscribe: update => cycles.subscribeToList(project, update, JSON.parse(key) as CycleFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, cycles],
	})

	return state.status === 'ready' ? { status: 'ready', cycles: state.data } : state
}

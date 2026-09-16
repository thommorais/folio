import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Plan } from '_/core/domain/plan'
import type { PlanFilter } from '_/core/ports/plans'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type PlansState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly plans: readonly Plan[] }
	| { readonly status: 'failed'; readonly message: string }

export const usePlans = (project: string, filter?: PlanFilter): PlansState => {
	const { plans, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Plan>({
		load: () => plans.list(project, JSON.parse(key) as PlanFilter),
		subscribe: update => plans.subscribeToList(project, update, JSON.parse(key) as PlanFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, plans],
	})

	return state.status === 'ready' ? { status: 'ready', plans: state.data } : state
}

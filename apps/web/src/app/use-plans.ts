import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Plan } from '_/core/domain/plan'
import type { PlanFilter } from '_/core/ports/plans'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type PlansState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly plans: readonly Plan[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

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

	return state.status === Status.Ready ? { status: Status.Ready, plans: state.data } : state
}

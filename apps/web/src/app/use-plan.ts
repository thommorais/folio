import type { Plan } from '_/core/domain/plan'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type PlanState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly plan: Plan }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const usePlan = (project: string, id: string | undefined): PlanState => {
	const { plans, connection } = useContainer()

	const state = useLiveRecord<Plan>({
		load: () => plans.get(project, id ?? ''),
		subscribe: (recordId, onChange, onGone) => plans.subscribeToRecord(project, recordId, onChange, onGone),
		connection,
		deps: [project, id],
		skip: !id,
	})

	return state.status === 'ready' ? { status: 'ready', plan: state.data } : state
}

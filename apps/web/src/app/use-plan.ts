import type { Plan } from '_/core/domain/plan'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type PlanState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly plan: Plan }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const usePlan = (project: string, id: string | undefined): PlanState => {
	const { plans, connection } = useContainer()

	const state = useLiveRecord<Plan>({
		load: () => plans.get(project, id ?? ''),
		subscribe: (recordId, onChange, onGone) => plans.subscribeToRecord(project, recordId, onChange, onGone),
		connection,
		deps: [project, id],
		skip: !id,
	})

	return state.status === Status.Ready ? { status: Status.Ready, plan: state.data } : state
}

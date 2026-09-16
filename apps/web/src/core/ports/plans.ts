import type { Sort, PlanSortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Plan, PlanStatus } from '../domain/plan'
import type { Unsubscribe } from './subscription'

export type PlanFilter = {
	readonly sort?: Sort<PlanSortField>
	readonly ticketId?: string
	readonly status?: readonly PlanStatus[]
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type PlansPort = {
	readonly count: (project: string, filter?: PlanFilter) => Promise<Result<number>>
	readonly get: (project: string, id: string) => Promise<Result<Plan>>
	readonly list: (project: string, filter?: PlanFilter) => Promise<Result<ReadonlyArray<Plan>>>
	readonly subscribeToList: (
		project: string,
		update: (plan: Plan, action: ActionEvent) => void,
		filter?: PlanFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Plan) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

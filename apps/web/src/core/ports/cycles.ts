import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Cycle } from '../domain/cycle'
import type { Unsubscribe } from './subscription'

export type CycleFilter = {
	readonly ticketId?: string
	readonly limit?: number
	readonly offset?: number
}

export type CyclesPort = {
	readonly list: (project: string, filter?: CycleFilter) => Promise<Result<ReadonlyArray<Cycle>>>
	readonly subscribeToList: (
		project: string,
		update: (cycle: Cycle, action: ActionEvent) => void,
		filter?: CycleFilter,
	) => Promise<Result<Unsubscribe>>
}

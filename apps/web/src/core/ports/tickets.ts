import type { Sort, TicketSortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Priority } from '../domain/todo'
import type { Ticket, TicketStatus } from '../domain/ticket'
import type { Unsubscribe } from './subscription'

export type TicketFilter = {
	readonly sort?: Sort<TicketSortField>
	readonly parentId?: string
	readonly status?: readonly TicketStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type TicketsPort = {
	readonly count: (project: string, filter?: TicketFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: TicketFilter) => Promise<Result<ReadonlyArray<Ticket>>>
	readonly get: (project: string, slug: string) => Promise<Result<Ticket>>
	readonly subscribeToList: (
		project: string,
		update: (ticket: Ticket, action: ActionEvent) => void,
		filter?: TicketFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Ticket) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

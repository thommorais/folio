import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { PlanLog, TicketLog, TodoLog } from '../domain/worklog'
import type { Unsubscribe } from './subscription'

export type TicketLogFilter = {
	readonly ticketId?: string
	readonly cycleId?: string
	readonly limit?: number
	readonly offset?: number
}

export type PlanLogFilter = {
	readonly planId?: string
	readonly limit?: number
	readonly offset?: number
}

export type TodoLogFilter = {
	readonly todoId?: string
	readonly limit?: number
	readonly offset?: number
}

export type WorkLogsPort = {
	readonly listTicketLogs: (project: string, filter?: TicketLogFilter) => Promise<Result<ReadonlyArray<TicketLog>>>
	readonly listPlanLogs: (project: string, filter?: PlanLogFilter) => Promise<Result<ReadonlyArray<PlanLog>>>
	readonly listTodoLogs: (project: string, filter?: TodoLogFilter) => Promise<Result<ReadonlyArray<TodoLog>>>
	readonly subscribeToTicketLogs: (
		project: string,
		update: (entry: TicketLog, action: ActionEvent) => void,
		filter?: TicketLogFilter,
	) => Promise<Result<Unsubscribe>>
}

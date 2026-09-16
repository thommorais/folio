import type { Branded } from './branded'
import type { CycleId } from './cycle'
import type { PlanId } from './plan'
import type { ProjectId, UserId } from './project'
import type { TicketId } from './ticket'
import type { TodoId } from './todo'

export type WorkLogId = Branded<string, 'WorkLogId'>

export const workLogId = (value: string): WorkLogId => value as WorkLogId

type Base = {
	readonly id: WorkLogId
	readonly projectId: ProjectId
	readonly body: string
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

export type TicketLog = Base & {
	readonly ticketId: TicketId
	readonly cycleId: CycleId | undefined
}

export type PlanLog = Base & {
	readonly planId: PlanId
}

export type TodoLog = Base & {
	readonly todoId: TodoId
}

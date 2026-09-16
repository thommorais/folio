import type { Branded } from './branded'
import type { PlanId } from './plan'
import type { ProjectId, UserId } from './project'
import type { TicketId } from './ticket'
import type { TodoId } from './todo'

export type LogId = Branded<string, 'LogId'>

export const logId = (value: string): LogId => value as LogId

export type JournalEntry = {
	readonly id: LogId
	readonly projectId: ProjectId
	readonly ticketId: TicketId | undefined
	readonly planId: PlanId | undefined
	readonly todoId: TodoId | undefined
	readonly slug: string
	readonly title: string
	readonly body: string
	readonly branch: string
	readonly pr: string
	readonly externalRef: string
	readonly tags: readonly string[]
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

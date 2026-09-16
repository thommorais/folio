import type { Branded } from './branded'
import type { PlanId } from './plan'
import type { TicketId } from './ticket'
import type { ProjectId, UserId } from './project'

export type TodoId = Branded<string, 'TodoId'>

export const todoId = (value: string): TodoId => value as TodoId

export const TODO_STATUSES = ['pending', 'in_progress', 'done', 'blocked', 'cancelled'] as const

export type TodoStatus = (typeof TODO_STATUSES)[number]

export const PRIORITIES = ['low', 'medium', 'high'] as const

export type Priority = (typeof PRIORITIES)[number]

export type Todo = {
	readonly id: TodoId
	readonly projectId: ProjectId
	readonly ticketId: TicketId | undefined
	readonly planId: PlanId | undefined
	readonly title: string
	readonly details: string
	readonly status: TodoStatus
	readonly priority: Priority
	readonly tags: readonly string[]
	readonly position: number
	readonly dependsOn: readonly TodoId[]
	readonly dueDate: Date | undefined
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

export const isTerminal = (status: TodoStatus): boolean => status === 'done' || status === 'cancelled'

import { cycleId as toCycleId } from '_/core/domain/cycle'
import { planId as toPlanId } from '_/core/domain/plan'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import { todoId as toTodoId } from '_/core/domain/todo'
import type { PlanLog, TicketLog, TodoLog } from '_/core/domain/worklog'
import { workLogId as toWorkLogId } from '_/core/domain/worklog'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { TicketLogFilter, WorkLogsPort } from '_/core/ports/worklogs'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import {
	Collections,
	type JournPlanLogsResponse,
	type JournTicketLogsResponse,
	type JournTodoLogsResponse,
} from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type TicketLogRecord = JournTicketLogsResponse
type PlanLogRecord = JournPlanLogsResponse
type TodoLogRecord = JournTodoLogsResponse

type TicketLogColumns = {
	'project.slug': string
	ticket: string
	cycle: string
}

type PlanLogColumns = {
	'project.slug': string
	plan: string
}

type TodoLogColumns = {
	'project.slug': string
	todo: string
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toTicketLog = (record: TicketLogRecord): TicketLog => ({
	id: toWorkLogId(record.id),
	projectId: toProjectId(record.project),
	ticketId: toTicketId(record.ticket),
	cycleId: record.cycle ? toCycleId(record.cycle) : undefined,
	body: record.body ?? '',
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const toPlanLog = (record: PlanLogRecord): PlanLog => ({
	id: toWorkLogId(record.id),
	projectId: toProjectId(record.project),
	planId: toPlanId(record.plan),
	body: record.body ?? '',
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const toTodoLog = (record: TodoLogRecord): TodoLog => ({
	id: toWorkLogId(record.id),
	projectId: toProjectId(record.project),
	todoId: toTodoId(record.todo),
	body: record.body ?? '',
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

export const createWorkLogsAdapter = (): WorkLogsPort => {
	const client = getPocketBaseClient()
	const ticketLogs = client.collection(Collections.JournTicketLogs)
	const planLogs = client.collection(Collections.JournPlanLogs)
	const todoLogs = client.collection(Collections.JournTodoLogs)

	const ticketColumns = (project: string, filter: TicketLogFilter) =>
		filterFor<TicketLogColumns>()([
			{ field: 'project.slug', comparator: 'eq', value: project },
			{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
			{ field: 'cycle', comparator: 'eq', value: filter.cycleId },
		])

	return {
		listTicketLogs: async (project, filter = {}): Promise<Result<readonly TicketLog[]>> => {
			const { expr, params } = ticketColumns(project, filter)

			const { data, error } = await tryCatch(
				paginate<TicketLogRecord>(ticketLogs, filter, {
					filter: client.filter(expr, params),
					sort: '-created',
				}),
			)

			return error
				? err(new Error(`Failed to list ticket work logs: ${error.message}`, { cause: error }))
				: ok(data.map(toTicketLog))
		},

		listPlanLogs: async (project, filter = {}): Promise<Result<readonly PlanLog[]>> => {
			const { expr, params } = filterFor<PlanLogColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'plan', comparator: 'eq', value: filter.planId },
			])

			const { data, error } = await tryCatch(
				paginate<PlanLogRecord>(planLogs, filter, {
					filter: client.filter(expr, params),
					sort: '-created',
				}),
			)

			return error
				? err(new Error(`Failed to list plan work logs: ${error.message}`, { cause: error }))
				: ok(data.map(toPlanLog))
		},

		listTodoLogs: async (project, filter = {}): Promise<Result<readonly TodoLog[]>> => {
			const { expr, params } = filterFor<TodoLogColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'todo', comparator: 'eq', value: filter.todoId },
			])

			const { data, error } = await tryCatch(
				paginate<TodoLogRecord>(todoLogs, filter, {
					filter: client.filter(expr, params),
					sort: '-created',
				}),
			)

			return error
				? err(new Error(`Failed to list todo work logs: ${error.message}`, { cause: error }))
				: ok(data.map(toTodoLog))
		},

		subscribeToTicketLogs: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = ticketColumns(project, filter)

			try {
				const unsubscribe = await ticketLogs.subscribe<TicketLogRecord>(
					'*',
					event => {
						update(toTicketLog(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to ticket work logs: ${message(error)}`))
			}
		},
	}
}

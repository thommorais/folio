import { sortExpr } from './sort'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import type { Priority, Todo, TodoStatus } from '_/core/domain/todo'
import { planId as toPlanId } from '_/core/domain/plan'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import { todoId as toTodoId } from '_/core/domain/todo'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { TodoFilter, TodosPort } from '_/core/ports/todos'
import type { ActionEvent } from '_/types'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournTodosResponse } from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { filterFor } from './filter-builder'
import { countRows } from './count-rows'
import { paginate } from './paginate'

type TodoRecord = JournTodosResponse<string[], string[]>

type TodoColumns = {
	'project.slug': string
	id: string
	title: string
	// The stored columns are "ticket" and "plan"; the filter names them
	// ticketId and planId to match the domain, so the two cannot simply be
	// intersected.
	ticket: string
	plan: string
} & Omit<TodoFilter, 'ticketId' | 'planId' | 'sort'>

const toTodo = (record: TodoRecord): Todo => ({
	id: toTodoId(record.id),
	projectId: toProjectId(record.project),
	ticketId: record.ticket ? toTicketId(record.ticket) : undefined,
	planId: record.plan ? toPlanId(record.plan) : undefined,
	title: record.title,
	details: record.details ?? '',
	status: record.status as TodoStatus,
	priority: record.priority as Priority,
	tags: record.tags ?? [],
	position: record.position ?? 0,
	dependsOn: (record.depends_on ?? []).map(toTodoId),
	dueDate: record.due_date ? new Date(record.due_date) : undefined,
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const columns = (project: string, filter: TodoFilter) =>
	filterFor<TodoColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
		{ field: 'plan', comparator: 'eq', value: filter.planId },
		{ field: 'status', comparator: 'anyOf', value: filter.status },
		{ field: 'priority', comparator: 'eq', value: filter.priority },
		{ field: 'tags', comparator: 'containsAll', value: filter.tags },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

export const createTodosAdapter = (): TodosPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournTodos)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection, { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count todos: ${error.message}`, { cause: error })) : ok(data)
		},

		get: async (project, id): Promise<Result<Todo>> => {
			const { expr, params } = filterFor<TodoColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'id', comparator: 'eq', value: id },
			])

			const { data, error } = await tryCatch(collection.getFirstListItem<TodoRecord>(client.filter(expr, params)))

			return error ? err(new Error(`Failed to load todo ${id}: ${error.message}`, { cause: error })) : ok(toTodo(data))
		},

		list: async (project, filter = {}): Promise<Result<readonly Todo[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<TodoRecord>(collection, filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, 'position'),
				}),
			)

			return error ? err(new Error(`Failed to list todos: ${error.message}`, { cause: error })) : ok(data.map(toTodo))
		},
		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			try {
				const { expr, params } = columns(project, filter)

				const unsubscribe = await collection.subscribe<TodoRecord>(
					'*',
					e => {
						update(toTodo(e.record), e.action as ActionEvent)
					},
					{
						filter: client.filter(expr, params),
						sort: sortExpr(filter.sort, 'position'),
					},
				)
				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to list: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(collection, id, toTodo, onChange, onGone, 'todo'),
	}
}

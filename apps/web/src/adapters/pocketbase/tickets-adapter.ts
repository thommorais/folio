import { sortExpr } from './sort'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import type { Ticket, TicketStatus, WayfinderType } from '_/core/domain/ticket'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import type { Priority } from '_/core/domain/todo'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { TicketFilter, TicketsPort } from '_/core/ports/tickets'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournTicketsResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { countRows } from './count-rows'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type TicketRecord = JournTicketsResponse<string[], string[]>

type TicketColumns = {
	'project.slug': string
	parent: string
	slug: string
	title: string
	body: string
	status: TicketStatus
	priority: Priority
	tags: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toTicket = (record: TicketRecord): Ticket => ({
	id: toTicketId(record.id),
	projectId: toProjectId(record.project),
	parentId: record.parent ? toTicketId(record.parent) : undefined,
	slug: record.slug,
	title: record.title,
	body: record.body ?? '',
	status: record.status as TicketStatus,
	priority: record.priority as Priority,
	assignee: record.assignee ? toUserId(record.assignee) : undefined,
	tags: record.tags ?? [],
	externalRef: record.external_ref ?? '',
	dependsOn: (record.depends_on ?? []).map(toTicketId),
	wayfinder: record.wayfinder ? (record.wayfinder as WayfinderType) : undefined,
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const columns = (project: string, filter: TicketFilter) =>
	filterFor<TicketColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'parent', comparator: 'eq', value: filter.parentId },
		{ field: 'status', comparator: 'anyOf', value: filter.status },
		{ field: 'priority', comparator: 'eq', value: filter.priority },
		{ field: 'tags', comparator: 'containsAll', value: filter.tags },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

export const createTicketsAdapter = (): TicketsPort => {
	const client = getPocketBaseClient()
	const collection = () => client.collection(Collections.JournTickets)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection(), { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count tickets: ${error.message}`, { cause: error })) : ok(data)
		},

		list: async (project, filter = {}): Promise<Result<readonly Ticket[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<TicketRecord>(collection(), filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, '-created'),
				}),
			)

			return error
				? err(new Error(`Failed to list tickets: ${error.message}`, { cause: error }))
				: ok(data.map(toTicket))
		},

		get: async (project, slug): Promise<Result<Ticket>> => {
			// A slug is unique only within a project, so both halves are bound.
			const { expr, params } = filterFor<TicketColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'slug', comparator: 'eq', value: slug },
			])

			const { data, error } = await tryCatch(collection().getFirstListItem<TicketRecord>(client.filter(expr, params)))

			return error
				? err(new Error(`Failed to load ticket ${slug}: ${error.message}`, { cause: error }))
				: ok(toTicket(data))
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection().subscribe<TicketRecord>(
					'*',
					event => {
						update(toTicket(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to tickets: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(collection(), id, toTicket, onChange, onGone, 'ticket'),
	}
}

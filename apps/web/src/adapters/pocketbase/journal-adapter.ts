import { sortExpr } from './sort'
import type { JournalEntry } from '_/core/domain/journal'
import { logId as toLogId } from '_/core/domain/journal'
import { planId as toPlanId } from '_/core/domain/plan'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import { todoId as toTodoId } from '_/core/domain/todo'
import type { JournalFilter, JournalPort } from '_/core/ports/journal'
import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournJournalResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { filterFor } from './filter-builder'
import { countRows } from './count-rows'
import { paginate } from './paginate'

type JournalRecord = JournJournalResponse<unknown, string[]>

type LogColumns = {
	'project.slug': string
	slug: string
	title: string
	body: string
	branch: string
	ticket: string
	external_ref: string
	tags: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toJournalEntry = (record: JournalRecord): JournalEntry => ({
	id: toLogId(record.id),
	projectId: toProjectId(record.project),
	ticketId: record.ticket ? toTicketId(record.ticket) : undefined,
	planId: record.plan ? toPlanId(record.plan) : undefined,
	todoId: record.todo ? toTodoId(record.todo) : undefined,
	slug: record.slug,
	title: record.title,
	body: record.body ?? '',
	branch: record.branch ?? '',
	pr: record.pr ?? '',
	externalRef: record.external_ref ?? '',
	tags: record.tags ?? [],
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const columns = (project: string, filter: JournalFilter) =>
	filterFor<LogColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
		{ field: 'branch', comparator: 'eq', value: filter.branch },
		{ field: 'external_ref', comparator: 'eq', value: filter.externalRef },
		{ field: 'tags', comparator: 'containsAll', value: filter.tags },
		{ field: 'title', comparator: 'contains', value: filter.search },
		{ field: 'created', comparator: 'gte', value: filter.since },
		{ field: 'created', comparator: 'lte', value: filter.until },
	])

export const createJournalAdapter = (): JournalPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournJournal)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection, { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count logs: ${error.message}`, { cause: error })) : ok(data)
		},

		list: async (project, filter = {}): Promise<Result<readonly JournalEntry[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<JournalRecord>(collection, filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, '-created'),
				}),
			)

			return error
				? err(new Error(`Failed to list logs: ${error.message}`, { cause: error }))
				: ok(data.map(toJournalEntry))
		},

		get: async (project, slug): Promise<Result<JournalEntry>> => {
			// A slug is unique only within a project, so both halves are bound.
			const { expr, params } = filterFor<LogColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'slug', comparator: 'eq', value: slug },
			])

			const { data, error } = await tryCatch(collection.getFirstListItem<JournalRecord>(client.filter(expr, params)))

			return error
				? err(new Error(`Failed to load journal entry ${slug}: ${error.message}`, { cause: error }))
				: ok(toJournalEntry(data))
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection.subscribe<JournalRecord>(
					'*',
					event => {
						update(toJournalEntry(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to logs: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(collection, id, toJournalEntry, onChange, onGone, 'journal entry'),
	}
}

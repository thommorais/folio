import { sortExpr } from './sort'
import type { Doc } from '_/core/domain/doc'
import { docId as toDocId } from '_/core/domain/doc'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import type { DocFilter, DocsPort } from '_/core/ports/docs'
import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournDocsResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { filterFor } from './filter-builder'
import { countRows } from './count-rows'
import { paginate } from './paginate'

type DocRecord = JournDocsResponse<string[]>

type DocColumns = {
	'project.slug': string
	ticket: string
	slug: string
	title: string
	body: string
	tags: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toDoc = (record: DocRecord): Doc => ({
	id: toDocId(record.id),
	projectId: toProjectId(record.project),
	ticketId: record.ticket ? toTicketId(record.ticket) : undefined,
	slug: record.slug,
	title: record.title,
	body: record.body ?? '',
	tags: record.tags ?? [],
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const columns = (project: string, filter: DocFilter) =>
	filterFor<DocColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
		{ field: 'tags', comparator: 'containsAll', value: filter.tags },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

export const createDocsAdapter = (): DocsPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournDocs)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection, { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count docs: ${error.message}`, { cause: error })) : ok(data)
		},

		list: async (project, filter = {}): Promise<Result<readonly Doc[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<DocRecord>(collection, filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, 'slug'),
				}),
			)

			return error ? err(new Error(`Failed to list docs: ${error.message}`, { cause: error })) : ok(data.map(toDoc))
		},

		get: async (project, slug): Promise<Result<Doc>> => {
			// A slug is unique only within a project, so both halves are bound.
			const { expr, params } = filterFor<DocColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'slug', comparator: 'eq', value: slug },
			])

			const { data, error } = await tryCatch(collection.getFirstListItem<DocRecord>(client.filter(expr, params)))

			return error ? err(new Error(`Failed to load doc ${slug}: ${error.message}`, { cause: error })) : ok(toDoc(data))
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection.subscribe<DocRecord>(
					'*',
					event => {
						update(toDoc(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to docs: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(collection, id, toDoc, onChange, onGone, 'doc'),
	}
}

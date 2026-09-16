import type { Cycle, Phase } from '_/core/domain/cycle'
import { cycleId as toCycleId } from '_/core/domain/cycle'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import type { CycleFilter, CyclesPort } from '_/core/ports/cycles'
import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournCyclesResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type CycleRecord = JournCyclesResponse

type CycleColumns = {
	'project.slug': string
	ticket: string
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toCycle = (record: CycleRecord): Cycle => ({
	id: toCycleId(record.id),
	projectId: toProjectId(record.project),
	ticketId: toTicketId(record.ticket),
	ordinal: record.ordinal ?? 0,
	phase: record.phase as Phase,
	resolution: record.resolution ?? '',
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
	closedAt: record.closed_at ? new Date(record.closed_at) : undefined,
})

const columns = (project: string, filter: CycleFilter) =>
	filterFor<CycleColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
	])

export const createCyclesAdapter = (): CyclesPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournCycles)

	return {
		list: async (project, filter = {}): Promise<Result<readonly Cycle[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<CycleRecord>(collection, filter, {
					filter: client.filter(expr, params),
					sort: 'ordinal',
				}),
			)

			return error ? err(new Error(`Failed to list cycles: ${error.message}`, { cause: error })) : ok(data.map(toCycle))
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection.subscribe<CycleRecord>(
					'*',
					event => {
						update(toCycle(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to cycles: ${message(error)}`))
			}
		},
	}
}

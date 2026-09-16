import { sortExpr } from './sort'
import type { Plan, PlanStatus } from '_/core/domain/plan'
import { planId as toPlanId } from '_/core/domain/plan'
import { ticketId as toTicketId } from '_/core/domain/ticket'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import type { PlanFilter, PlansPort } from '_/core/ports/plans'
import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournPlansResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { filterFor } from './filter-builder'
import { countRows } from './count-rows'
import { paginate } from './paginate'

type PlanRecord = JournPlansResponse<string[]>

type PlanColumns = {
	'project.slug': string
	id: string
	ticket: string
	title: string
	goal: string
	status: PlanStatus
	tags: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toPlan = (record: PlanRecord): Plan => ({
	id: toPlanId(record.id),
	projectId: toProjectId(record.project),
	ticketId: record.ticket ? toTicketId(record.ticket) : undefined,
	title: record.title,
	goal: record.goal ?? '',
	status: record.status as PlanStatus,
	tags: record.tags ?? [],
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const columns = (project: string, filter: PlanFilter) =>
	filterFor<PlanColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'ticket', comparator: 'eq', value: filter.ticketId },
		{ field: 'status', comparator: 'anyOf', value: filter.status },
		{ field: 'tags', comparator: 'containsAll', value: filter.tags },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

export const createPlansAdapter = (): PlansPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournPlans)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection, { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count plans: ${error.message}`, { cause: error })) : ok(data)
		},

		get: async (project, id): Promise<Result<Plan>> => {
			const { expr, params } = filterFor<PlanColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'id', comparator: 'eq', value: id },
			])

			const { data, error } = await tryCatch(collection.getFirstListItem<PlanRecord>(client.filter(expr, params)))

			return error ? err(new Error(`Failed to load plan ${id}: ${error.message}`, { cause: error })) : ok(toPlan(data))
		},

		list: async (project, filter = {}): Promise<Result<readonly Plan[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<PlanRecord>(collection, filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, '-created'),
				}),
			)

			return error ? err(new Error(`Failed to list plans: ${error.message}`, { cause: error })) : ok(data.map(toPlan))
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection.subscribe<PlanRecord>(
					'*',
					event => {
						update(toPlan(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to plans: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(collection, id, toPlan, onChange, onGone, 'plan'),
	}
}

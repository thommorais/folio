import type { Client } from '_/core/domain/client'
import { clientId as toClientId } from '_/core/domain/client'
import type { ClientFilter, ClientsPort } from '_/core/ports/clients'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { ActionEvent } from '_/types'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournClientsResponse } from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type ClientColumns = {
	slug: string
	name: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const columns = (filter: ClientFilter) =>
	filterFor<ClientColumns>()([{ field: 'name', comparator: 'contains', value: filter.search }])

const toClient = (record: JournClientsResponse): Client => ({
	id: toClientId(record.id),
	slug: record.slug,
	name: record.name,
	site: record.site ?? '',
	logo: record.logo ?? '',
	descr: record.descr ?? '',
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

export const createClientsAdapter = (): ClientsPort => {
	const client = getPocketBaseClient()
	const clients = () => client.collection(Collections.JournClients)

	return {
		list: async (filter: ClientFilter = {}): Promise<Result<readonly Client[]>> => {
			const { expr, params } = columns(filter)

			const { data, error } = await tryCatch(
				paginate<JournClientsResponse>(clients(), filter, {
					filter: client.filter(expr, params),
					sort: 'name',
				}),
			)

			return error
				? err(new Error(`Failed to list clients: ${error.message}`, { cause: error }))
				: ok(data.map(toClient))
		},

		get: async (ref: string): Promise<Result<Client>> => {
			const { expr, params } = filterFor<ClientColumns>()([{ field: 'slug', comparator: 'eq', value: ref }])

			const { data, error } = await tryCatch(
				clients().getFirstListItem<JournClientsResponse>(client.filter(expr, params)),
			)

			return error
				? err(new Error(`Failed to load client ${ref}: ${error.message}`, { cause: error }))
				: ok(toClient(data))
		},

		subscribeToList: async (update, filter: ClientFilter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(filter)

			try {
				const unsubscribe = await clients().subscribe<JournClientsResponse>(
					'*',
					event => {
						update(toClient(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to clients: ${message(error)}`))
			}
		},
	}
}

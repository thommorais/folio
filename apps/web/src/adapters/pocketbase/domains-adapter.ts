import { clientId as toClientId } from '_/core/domain/client'
import type { Domain, Member, Role } from '_/core/domain/domain'
import { domainId as toDomainId } from '_/core/domain/domain'
import { userId as toUserId } from '_/core/domain/project'
import type { DomainFilter, DomainsPort } from '_/core/ports/domains'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { ActionEvent } from '_/types'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import {
	Collections,
	type JournClientsResponse,
	type JournDomainsResponse,
	type JournMembersResponse,
	type UsersResponse,
} from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type MemberRecord = JournMembersResponse<{ user?: UsersResponse }>

type DomainRecord = JournDomainsResponse<{
	client?: JournClientsResponse
	journ_members_via_domain?: MemberRecord[]
}>

type DomainColumns = {
	client: string
	slug: string
	name: string
	created: Date
	updated: Date
}

const MEMBER_EXPAND = 'journ_members_via_domain.user,client'

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toMember = (record: MemberRecord): Member => ({
	userId: toUserId(record.user),
	role: record.role as Role,
	email: record.expand?.user?.email ?? '',
	name: record.expand?.user?.name ?? '',
})

const columns = (filter: DomainFilter) =>
	filterFor<DomainColumns>()([
		{ field: 'client', comparator: 'eq', value: filter.clientId },
		{ field: 'name', comparator: 'contains', value: filter.search },
	])

const toDomain = (record: DomainRecord): Domain => ({
	id: toDomainId(record.id),
	clientId: toClientId(record.client),
	slug: record.slug,
	name: record.name,
	descr: record.descr ?? '',
	members: (record.expand?.journ_members_via_domain ?? []).map(toMember),
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

export const createDomainsAdapter = (): DomainsPort => {
	const client = getPocketBaseClient()
	const domains = () => client.collection(Collections.JournDomains)

	return {
		list: async (filter: DomainFilter = {}): Promise<Result<readonly Domain[]>> => {
			const { expr, params } = columns(filter)

			const { data, error } = await tryCatch(
				paginate<DomainRecord>(domains(), filter, {
					filter: client.filter(expr, params),
					expand: MEMBER_EXPAND,
					sort: 'name',
				}),
			)

			return error
				? err(new Error(`Failed to list domains: ${error.message}`, { cause: error }))
				: ok(data.map(toDomain))
		},

		get: async (clientRef: string, ref: string): Promise<Result<Domain>> => {
			const { expr, params } = filterFor<{ 'client.slug': string; slug: string }>()([
				{ field: 'client.slug', comparator: 'eq', value: clientRef },
				{ field: 'slug', comparator: 'eq', value: ref },
			])

			const { data, error } = await tryCatch(
				domains().getFirstListItem<DomainRecord>(client.filter(expr, params), { expand: MEMBER_EXPAND }),
			)

			return error
				? err(new Error(`Failed to load domain ${clientRef}/${ref}: ${error.message}`, { cause: error }))
				: ok(toDomain(data))
		},

		subscribeToList: async (update, filter: DomainFilter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(filter)

			try {
				const unsubscribe = await domains().subscribe<DomainRecord>(
					'*',
					event => {
						update(toDomain(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params), expand: MEMBER_EXPAND },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to domains: ${message(error)}`))
			}
		},
	}
}

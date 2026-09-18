import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Domain } from '../domain/domain'
import type { Unsubscribe } from './subscription'

export type DomainFilter = {
	readonly clientId?: string
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type DomainsPort = {
	readonly list: (filter?: DomainFilter) => Promise<Result<readonly Domain[]>>
	// A domain slug is unique within its client, so both halves address one.
	readonly get: (client: string, ref: string) => Promise<Result<Domain>>
	readonly subscribeToList: (
		update: (domain: Domain, action: ActionEvent) => void,
		filter?: DomainFilter,
	) => Promise<Result<Unsubscribe>>
}

import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Client } from '../domain/client'
import type { Unsubscribe } from './subscription'

export type ClientFilter = {
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type ClientsPort = {
	readonly list: (filter?: ClientFilter) => Promise<Result<readonly Client[]>>
	readonly get: (ref: string) => Promise<Result<Client>>
	readonly subscribeToList: (
		update: (client: Client, action: ActionEvent) => void,
		filter?: ClientFilter,
	) => Promise<Result<Unsubscribe>>
}

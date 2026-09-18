import type { Branded } from './branded'

export type ClientId = Branded<string, 'ClientId'>

export const clientId = (value: string): ClientId => value as ClientId

export type Client = {
	readonly id: ClientId
	readonly slug: string
	readonly name: string
	readonly site: string
	readonly logo: string
	readonly descr: string
	readonly createdAt: Date
	readonly updatedAt: Date
}

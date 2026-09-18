import type { Branded } from './branded'
import type { DomainId } from './domain'

export type ProjectId = Branded<string, 'ProjectId'>
export type UserId = Branded<string, 'UserId'>

export const projectId = (value: string): ProjectId => value as ProjectId
export const userId = (value: string): UserId => value as UserId

export type Project = {
	readonly id: ProjectId
	readonly domainId: DomainId
	readonly slug: string
	readonly name: string
	readonly descr: string
	readonly archived: boolean
	readonly createdAt: Date
	readonly updatedAt: Date
}

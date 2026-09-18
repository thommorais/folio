import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Project } from '../domain/project'
import type { Unsubscribe } from './subscription'

export type ProjectFilter = {
	readonly domainId?: string
	readonly includeArchived?: boolean
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type ProjectsPort = {
	readonly list: (filter?: ProjectFilter) => Promise<Result<readonly Project[]>>
	readonly get: (ref: string) => Promise<Result<Project>>
	readonly subscribeToList: (
		update: (project: Project, action: ActionEvent) => void,
		filter?: ProjectFilter,
	) => Promise<Result<Unsubscribe>>
}

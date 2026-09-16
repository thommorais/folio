import type { Result } from '_/lib/result'
import type { ProjectId } from '../domain/project'

export const SEARCH_KINDS = ['log', 'doc', 'todo', 'plan'] as const

export type SearchKind = (typeof SEARCH_KINDS)[number]

export type SearchHit = {
	readonly kind: SearchKind
	readonly id: string
	readonly projectId: ProjectId
	readonly projectSlug: string
	readonly title: string
	readonly snippet: string
	readonly tags: readonly string[]
	readonly createdAt: Date
}

export type SearchQuery = {
	readonly text: string
	readonly kinds?: readonly SearchKind[]
	readonly limit?: number
}

export type SearchPort = {
	readonly search: (query: SearchQuery) => Promise<Result<readonly SearchHit[]>>
}

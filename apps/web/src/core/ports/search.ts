import type { Result } from '_/lib/result'
import type { ProjectId } from '../domain/project'

// These are the API's own kind names, so a hit needs no translation at the
// boundary. 'journal' was called 'log' while the client searched collections
// directly and had to pick its own vocabulary.
export const SEARCH_KINDS = ['journal', 'doc', 'todo', 'plan', 'ticket', 'worklog', 'resolution'] as const

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

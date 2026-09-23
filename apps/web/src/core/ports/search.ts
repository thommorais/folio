import type { Result } from '_/lib/result'
import type { ProjectId } from '../domain/project'

// The display order in the palette, and the set a query asks for. The names
// match the API's kinds, so a hit needs no translation.
//
// The API also answers 'worklog' and 'resolution'. Neither is addressable in
// this app, so the palette does not ask for them rather than offering a row
// that cannot be opened.
export const SEARCH_KINDS = ['knowledge', 'ticket', 'todo', 'plan', 'doc', 'journal'] as const

export type SearchKind = (typeof SEARCH_KINDS)[number]

export type SearchHit = {
	readonly kind: SearchKind
	readonly id: string
	readonly projectId: ProjectId
	// Empty for a hit that belongs to no project, which only knowledge can be.
	readonly projectSlug: string
	readonly clientSlug: string
	readonly domainSlug: string
	readonly slug: string
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

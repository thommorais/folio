import type { Result } from '_/lib/result'

// Knowledge is durable, reusable notes: a tip, a snippet, a fix worth keeping.
// It is the one resource with no project, so unlike every other port here
// nothing it takes is scoped to one.
export type Knowledge = {
	readonly id: string
	readonly slug: string
	readonly title: string
	readonly body: string
	// Empty when the note belongs to no project, which is the common case. It
	// records where the note was learned and does not restrict who may read it.
	readonly projectId: string
	readonly tags: readonly string[]
	readonly createdAt: Date
	readonly updatedAt: Date
}

export type KnowledgeFilter = {
	readonly text?: string
	readonly tags?: readonly string[]
	readonly limit?: number
}

export type KnowledgePort = {
	readonly list: (filter?: KnowledgeFilter) => Promise<Result<readonly Knowledge[]>>
	// Takes an id or a slug: the slug namespace is global, so no project is
	// needed to resolve one.
	readonly get: (ref: string) => Promise<Result<Knowledge>>
}

import type { Knowledge, KnowledgeFilter, KnowledgePort } from '_/core/ports/knowledge'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { getPocketBaseClient } from './client'

const DEFAULT_LIMIT = 50

// Knowledge is reached through the folio API rather than the collection, so
// that the permission rule lives in one place. send() carries the PocketBase
// auth token, which is the token the folio API authenticates with.
type KnowledgeResponse = {
	id: string
	slug: string
	title: string
	body?: string
	project_id?: string
	tags: string[]
	created_at: string
	updated_at: string
}

const toKnowledge = (note: KnowledgeResponse): Knowledge => ({
	id: note.id,
	slug: note.slug,
	title: note.title,
	body: note.body ?? '',
	projectId: note.project_id ?? '',
	tags: note.tags ?? [],
	createdAt: new Date(note.created_at),
	updatedAt: new Date(note.updated_at),
})

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

export const createKnowledgeAdapter = (): KnowledgePort => {
	const client = getPocketBaseClient()

	return {
		list: async (filter: KnowledgeFilter = {}): Promise<Result<readonly Knowledge[]>> => {
			const params = new URLSearchParams({ limit: String(filter.limit ?? DEFAULT_LIMIT) })
			if (filter.text?.trim()) params.set('q', filter.text.trim())
			if (filter.tags?.length) params.set('tags', filter.tags.join(','))

			const { data, error } = await tryCatch(
				client.send<{ knowledge: KnowledgeResponse[] }>(`/api/folio/knowledge?${params.toString()}`, {
					method: 'GET',
				}),
			)

			return error
				? err(new Error(`Failed to load knowledge: ${message(error)}`, { cause: error }))
				: ok((data.knowledge ?? []).map(toKnowledge))
		},

		get: async (ref: string): Promise<Result<Knowledge>> => {
			const { data, error } = await tryCatch(
				client.send<KnowledgeResponse>(`/api/folio/knowledge/${encodeURIComponent(ref)}`, { method: 'GET' }),
			)

			return error
				? err(new Error(`Failed to load note ${ref}: ${message(error)}`, { cause: error }))
				: ok(toKnowledge(data))
		},
	}
}

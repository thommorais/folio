import { projectId as toProjectId } from '_/core/domain/project'
import type { SearchHit, SearchKind, SearchPort, SearchQuery } from '_/core/ports/search'
import { SEARCH_KINDS } from '_/core/ports/search'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { getPocketBaseClient } from './client'

const DEFAULT_LIMIT = 20

// One ranked request to the folio API, rather than a query per kind merged by
// recency here. The API answers from an FTS5 index, so a title match outranks
// a body match and the scores are comparable across kinds, which merging
// separate queries could never be. It is also the only way to see knowledge,
// which belongs to no project and so has no collection this app can scope.
type SearchHitResponse = {
	kind: string
	id: string
	project_id: string
	project_slug: string
	domain_slug: string
	client_slug: string
	slug?: string
	title: string
	snippet?: string
	tags: string[]
	created_at: string
}

const toHit = (hit: SearchHitResponse): SearchHit => ({
	kind: hit.kind as SearchKind,
	id: hit.id,
	projectId: toProjectId(hit.project_id),
	projectSlug: hit.project_slug,
	clientSlug: hit.client_slug,
	domainSlug: hit.domain_slug,
	slug: hit.slug ?? '',
	title: hit.title,
	snippet: hit.snippet ?? '',
	tags: hit.tags ?? [],
	createdAt: new Date(hit.created_at),
})

export const createSearchAdapter = (): SearchPort => {
	const client = getPocketBaseClient()

	return {
		search: async ({ text, kinds, limit = DEFAULT_LIMIT }: SearchQuery): Promise<Result<readonly SearchHit[]>> => {
			const term = text.trim()
			if (term === '') return ok([])

			// The kinds are always sent, even unfiltered: the API answers more
			// of them than this app can open.
			const params = new URLSearchParams({
				q: term,
				limit: String(limit),
				kind: (kinds?.length ? kinds : SEARCH_KINDS).join(','),
			})

			// send() carries the PocketBase auth token, which is the same token
			// the folio API authenticates with.
			const { data, error } = await tryCatch(
				client.send<{ hits: SearchHitResponse[] }>(`/api/folio/search?${params.toString()}`, { method: 'GET' }),
			)

			return error
				? err(new Error(`Search failed: ${error.message}`, { cause: error }))
				: ok((data.hits ?? []).map(toHit))
		},
	}
}

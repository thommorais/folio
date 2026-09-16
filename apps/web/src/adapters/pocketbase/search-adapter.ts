import { projectId as toProjectId } from '_/core/domain/project'
import type { SearchHit, SearchKind, SearchPort, SearchQuery } from '_/core/ports/search'
import { SEARCH_KINDS } from '_/core/ports/search'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { ENVS } from '_/envs'
import { getPocketBaseClient } from './client'

const DEFAULT_LIMIT = 20

// The API's shape, which the client mirrors rather than reinterprets. Ranking
// and the snippet are computed server-side against the FTS5 index, so a hit
// arrives ready to render and must not be re-sorted here.
type SearchHitResponse = {
	readonly kind: string
	readonly id: string
	readonly project_id: string
	readonly project_slug: string
	readonly title: string
	readonly snippet?: string
	readonly tags: readonly string[] | null
	readonly created_at: string
}

const isKind = (value: string): value is SearchKind => (SEARCH_KINDS as readonly string[]).includes(value)

const toHit = (raw: SearchHitResponse): SearchHit | undefined => {
	if (!isKind(raw.kind)) return undefined

	return {
		kind: raw.kind,
		id: raw.id,
		projectId: toProjectId(raw.project_id),
		projectSlug: raw.project_slug,
		title: raw.title,
		snippet: raw.snippet ?? '',
		tags: raw.tags ?? [],
		createdAt: new Date(raw.created_at),
	}
}

export const createSearchAdapter = (): SearchPort => {
	const client = getPocketBaseClient()

	return {
		search: async ({ text, kinds, limit = DEFAULT_LIMIT }: SearchQuery): Promise<Result<readonly SearchHit[]>> => {
			const term = text.trim()
			if (term === '') return ok([])

			const params = new URLSearchParams({ q: term, limit: String(limit) })
			if (kinds?.length) params.set('kind', kinds.join(','))

			// Global rather than per-project: the palette has no project in
			// scope, and the endpoint already limits results to the caller's
			// memberships.
			const url = `${ENVS.PUBLIC_API_URL}/api/folio/search?${params.toString()}`

			const { data, error } = await tryCatch(
				fetch(url, { headers: { Authorization: client.authStore.token } }).then(async response => {
					if (!response.ok) {
						throw new Error(`${response.status} ${response.statusText}`)
					}
					return response.json() as Promise<{ hits?: readonly SearchHitResponse[] }>
				}),
			)

			if (error) {
				return err(new Error(`Search failed: ${error.message}`, { cause: error }))
			}

			// A kind the client does not know yet is skipped rather than
			// rendered as a broken row, so the API can add one first.
			return ok((data.hits ?? []).map(toHit).filter((hit): hit is SearchHit => hit !== undefined))
		},
	}
}

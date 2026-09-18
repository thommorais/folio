import { projectId as toProjectId } from '_/core/domain/project'
import type { SearchHit, SearchKind, SearchPort, SearchQuery } from '_/core/ports/search'
import { SEARCH_KINDS } from '_/core/ports/search'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import {
	Collections,
	type JournClientsResponse,
	type JournDomainsResponse,
	type JournProjectsResponse,
} from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'

const DEFAULT_LIMIT = 20

const SNIPPET_LENGTH = 200

type SearchableRecord = {
	id: string
	project: string
	slug?: string
	title: string
	body?: string
	goal?: string
	tags?: string[]
	created: string
	expand?: {
		project?: JournProjectsResponse<{
			domain?: JournDomainsResponse<{ client?: JournClientsResponse }>
		}>
	}
}

type SearchColumns = {
	title: string
	body: string
	goal: string
}

const sources: Record<SearchKind, { collection: string; text: keyof SearchColumns; kind?: string }> = {
	log: { collection: Collections.JournEntries, text: 'body', kind: 'journal' },
	doc: { collection: Collections.JournEntries, text: 'body', kind: 'doc' },
	todo: { collection: Collections.JournIssues, text: 'body', kind: 'todo' },
	plan: { collection: Collections.JournPlans, text: 'goal' },
}

// Mirrors rules.Snippet in the Go core: collapse whitespace, cut on a word
// boundary when one is near the limit.
const snippet = (text: string): string => {
	const clean = text.replace(/\s+/gu, ' ').trim()
	if (clean.length <= SNIPPET_LENGTH) return clean

	const cut = clean.slice(0, SNIPPET_LENGTH)
	const boundary = cut.lastIndexOf(' ')
	return `${(boundary > SNIPPET_LENGTH / 2 ? cut.slice(0, boundary) : cut).replace(/[,.;:\s]+$/u, '')}…`
}

const toHit = (kind: SearchKind, textField: keyof SearchColumns, record: SearchableRecord): SearchHit => ({
	kind,
	id: record.id,
	projectId: toProjectId(record.project),
	projectSlug: record.expand?.project?.slug ?? '',
	clientSlug: record.expand?.project?.expand?.domain?.expand?.client?.slug ?? '',
	domainSlug: record.expand?.project?.expand?.domain?.slug ?? '',
	slug: record.slug ?? '',
	title: record.title,
	snippet: snippet(String(record[textField as keyof SearchableRecord] ?? '')),
	tags: record.tags ?? [],
	createdAt: new Date(record.created),
})

export const createSearchAdapter = (): SearchPort => {
	const client = getPocketBaseClient()

	return {
		search: async ({ text, kinds, limit = DEFAULT_LIMIT }: SearchQuery): Promise<Result<readonly SearchHit[]>> => {
			const term = text.trim()
			if (term === '') return ok([])

			const wanted = kinds?.length ? kinds : SEARCH_KINDS

			const { data, error } = await tryCatch(
				Promise.all(
					wanted.map(async kind => {
						const { collection, text: textField, kind: rowKind } = sources[kind]
						const { expr, params } = filterFor<SearchColumns>()([
							{ field: 'title', comparator: 'contains', value: term },
							{ field: textField, comparator: 'contains', value: term },
						])

						// Both clauses target the same term, so OR them rather than AND.
						const matches = `(${expr.replace(' && ', ' || ')})`
						// Issues and entries each hold several kinds in one collection.
						const scoped = rowKind === undefined ? matches : `${matches} && kind = {:kind}`

						const { items } = await client.collection(collection).getList<SearchableRecord>(1, limit, {
							filter: client.filter(scoped, { ...params, kind: rowKind }),
							expand: 'project.domain.client',
							sort: '-created',
						})

						return items.map(record => toHit(kind, textField, record))
					}),
				),
			)

			return error
				? err(new Error(`Search failed: ${error.message}`, { cause: error }))
				: ok(
						data
							.flat()
							.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
							.slice(0, limit),
					)
		},
	}
}

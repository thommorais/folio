import { createFileRoute } from '@tanstack/react-router'
import { DOC_SORT_FIELDS, type DocSortField, type Sort } from '_/core/ports/sort'
import { Docs } from '_/pages/docs'
import { asSort, asString, asStrings } from '_/routes/search-params'

export type DocsSearch = {
	readonly ticket?: string
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<DocSortField>
}

export const Route = createFileRoute('/_authenticated/$slug/docs/')({
	validateSearch: (search: Record<string, unknown>): DocsSearch => ({
		ticket: asString(search.ticket),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(DOC_SORT_FIELDS, search.sort),
	}),
	component: Docs,
})

import { createFileRoute } from '@tanstack/react-router'
import { ENTRY_SORT_FIELDS, type EntrySortField, type Sort } from '_/core/ports/sort'
import { Logs } from '_/pages/journal'
import { asSort, asString, asStrings } from '_/routes/search-params'

export type LogsSearch = {
	readonly ticket?: string
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<EntrySortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/journal/')({
	validateSearch: (search: Record<string, unknown>): LogsSearch => ({
		ticket: asString(search.ticket),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(ENTRY_SORT_FIELDS, search.sort),
	}),
	component: Logs,
})

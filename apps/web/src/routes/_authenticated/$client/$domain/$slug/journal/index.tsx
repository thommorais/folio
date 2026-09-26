import { createFileRoute, redirect } from '@tanstack/react-router'
import { listMemory } from '_/adapters/browser/list-memory'
import { ADDRESSABLE_KINDS, type AddressableKind } from '_/core/domain/entry'
import { ENTRY_SORT_FIELDS, type EntrySortField, type Sort } from '_/core/ports/sort'
import { Journal } from '_/pages/journal'
import { asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type JournalSearch = {
	readonly kinds?: readonly AddressableKind[]
	readonly ticket?: string
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<EntrySortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/journal/')({
	beforeLoad: ({ search, cause, params }) => {
		if (cause !== 'enter') return
		const restored = listMemory.restore('/_authenticated/$client/$domain/$slug/journal/', search)
		if (restored) throw redirect({ to: '/$client/$domain/$slug/journal', params, search: restored })
	},
	validateSearch: (search: Record<string, unknown>): JournalSearch => ({
		kinds: asMembers(ADDRESSABLE_KINDS, search.kinds),
		ticket: asString(search.ticket),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(ENTRY_SORT_FIELDS, search.sort),
	}),
	component: Journal,
})

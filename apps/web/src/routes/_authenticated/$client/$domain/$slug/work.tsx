import { createFileRoute, redirect } from '@tanstack/react-router'
import { listMemory } from '_/adapters/browser/list-memory'
import { ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue'
import { WORK_SORT_FIELDS, type Sort, type WorkSortField } from '_/core/ports/sort'
import { Work } from '_/pages/work'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type WorkSearch = {
	readonly q?: string
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly sort?: Sort<WorkSortField>
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/work')({
	beforeLoad: ({ search, cause, params }) => {
		if (cause !== 'enter') return
		const restored = listMemory.restore('/_authenticated/$client/$domain/$slug/work', search)
		if (restored) throw redirect({ to: '/$client/$domain/$slug/work', params, search: restored })
	},
	validateSearch: (search: Record<string, unknown>): WorkSearch => ({
		q: asString(search.q),
		statuses: asMembers(ISSUE_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		sort: asSort(WORK_SORT_FIELDS, search.sort),
	}),
	component: Work,
})

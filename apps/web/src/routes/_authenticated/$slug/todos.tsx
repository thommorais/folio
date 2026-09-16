import { createFileRoute } from '@tanstack/react-router'
import { TODO_STATUSES, PRIORITIES, type Priority, type TodoStatus } from '_/core/domain/todo'
import { TODO_SORT_FIELDS, type Sort, type TodoSortField } from '_/core/ports/sort'
import { Todos } from '_/pages/todos'
import { asMember, asMembers, asSort, asString, asStrings } from '_/routes/search-params'

export type TodosSearch = {
	readonly todo?: string
	readonly ticket?: string
	readonly plan?: string
	readonly statuses?: readonly TodoStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<TodoSortField>
}

export const Route = createFileRoute('/_authenticated/$slug/todos')({
	validateSearch: (search: Record<string, unknown>): TodosSearch => ({
		todo: asString(search.todo),
		ticket: asString(search.ticket),
		plan: asString(search.plan),
		statuses: asMembers(TODO_STATUSES, search.statuses),
		priority: asMember(PRIORITIES, search.priority),
		tags: asStrings(search.tags),
		q: asString(search.q),
		sort: asSort(TODO_SORT_FIELDS, search.sort),
	}),
	component: Todos,
})

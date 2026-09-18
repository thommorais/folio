export type SortDirection = 'asc' | 'desc'

export type Sort<TField extends string> = {
	readonly field: TField
	readonly direction: SortDirection
}

export const ISSUE_SORT_FIELDS = ['position', 'title', 'status', 'priority', 'size', 'created', 'updated'] as const
export const ENTRY_SORT_FIELDS = ['title', 'slug', 'created', 'updated'] as const
export const TODO_SORT_FIELDS = ['position', 'title', 'status', 'priority', 'created', 'updated'] as const
export const PLAN_SORT_FIELDS = ['title', 'status', 'created', 'updated'] as const
export const TICKET_SORT_FIELDS = ['title', 'status', 'priority', 'created', 'updated'] as const
export const LOG_SORT_FIELDS = ['title', 'created', 'updated'] as const
export const DOC_SORT_FIELDS = ['title', 'slug', 'created', 'updated'] as const

export type IssueSortField = (typeof ISSUE_SORT_FIELDS)[number]
export type EntrySortField = (typeof ENTRY_SORT_FIELDS)[number]
export type TodoSortField = (typeof TODO_SORT_FIELDS)[number]
export type PlanSortField = (typeof PLAN_SORT_FIELDS)[number]
export type TicketSortField = (typeof TICKET_SORT_FIELDS)[number]
export type LogSortField = (typeof LOG_SORT_FIELDS)[number]
export type DocSortField = (typeof DOC_SORT_FIELDS)[number]

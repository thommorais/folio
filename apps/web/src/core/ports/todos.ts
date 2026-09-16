import type { Sort, TodoSortField } from './sort'
import type { Result } from '_/lib/result'
import type { Unsubscribe } from './subscription'
import type { ActionEvent } from '_/types'
import type { Priority, Todo, TodoStatus } from '../domain/todo'

export type TodoFilter = {
	readonly sort?: Sort<TodoSortField>
	readonly ticketId?: string
	readonly planId?: string
	readonly status?: readonly TodoStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type TodosPort = {
	readonly count: (project: string, filter?: TodoFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: TodoFilter) => Promise<Result<ReadonlyArray<Todo>>>
	// A todo carries no slug, so it is addressed by id the way the API
	// addresses one. The project is still bound, so an id from another project
	// cannot be reached by editing the URL.
	readonly get: (project: string, id: string) => Promise<Result<Todo>>
	readonly subscribeToList: (
		project: string,
		update: (todo: Todo, action: ActionEvent) => void,
		filter?: TodoFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Todo) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

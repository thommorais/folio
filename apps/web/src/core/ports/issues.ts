import type { Sort, IssueSortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Issue, IssueKind, IssueStatus, Priority } from '../domain/issue'
import type { Unsubscribe } from './subscription'

export type IssueFilter = {
	readonly sort?: Sort<IssueSortField>
	readonly kind?: IssueKind
	readonly parentId?: string
	readonly planId?: string
	readonly status?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type IssuesPort = {
	readonly count: (project: string, filter?: IssueFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: IssueFilter) => Promise<Result<ReadonlyArray<Issue>>>
	readonly get: (project: string, slug: string) => Promise<Result<Issue>>
	readonly getById: (project: string, id: string) => Promise<Result<Issue>>
	readonly subscribeToList: (
		project: string,
		update: (issue: Issue, action: ActionEvent) => void,
		filter?: IssueFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Issue) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

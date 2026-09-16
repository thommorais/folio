import type { Sort, DocSortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Doc } from '../domain/doc'
import type { Unsubscribe } from './subscription'

export type DocFilter = {
	readonly sort?: Sort<DocSortField>
	readonly ticketId?: string
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type DocsPort = {
	readonly count: (project: string, filter?: DocFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: DocFilter) => Promise<Result<ReadonlyArray<Doc>>>
	readonly get: (project: string, slug: string) => Promise<Result<Doc>>
	readonly subscribeToList: (
		project: string,
		update: (doc: Doc, action: ActionEvent) => void,
		filter?: DocFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Doc) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

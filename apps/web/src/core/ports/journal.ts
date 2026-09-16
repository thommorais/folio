import type { Sort, LogSortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { JournalEntry } from '../domain/journal'
import type { Unsubscribe } from './subscription'

export type JournalFilter = {
	readonly sort?: Sort<LogSortField>
	readonly ticketId?: string
	readonly branch?: string
	readonly externalRef?: string
	readonly tags?: readonly string[]
	readonly search?: string
	readonly since?: Date
	readonly until?: Date
	readonly limit?: number
	readonly offset?: number
}

export type JournalPort = {
	readonly count: (project: string, filter?: JournalFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: JournalFilter) => Promise<Result<ReadonlyArray<JournalEntry>>>
	readonly get: (project: string, slug: string) => Promise<Result<JournalEntry>>
	readonly subscribeToList: (
		project: string,
		update: (entry: JournalEntry, action: ActionEvent) => void,
		filter?: JournalFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: JournalEntry) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

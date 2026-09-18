import type { Sort, EntrySortField } from './sort'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import type { Entry, EntryKind } from '../domain/entry'
import type { Unsubscribe } from './subscription'

export type EntryFilter = {
	readonly sort?: Sort<EntrySortField>
	readonly kind?: EntryKind
	readonly issueId?: string
	readonly planId?: string
	readonly cycleId?: string
	readonly branch?: string
	readonly externalRef?: string
	readonly tags?: readonly string[]
	readonly search?: string
	readonly limit?: number
	readonly offset?: number
}

export type EntriesPort = {
	readonly count: (project: string, filter?: EntryFilter) => Promise<Result<number>>
	readonly list: (project: string, filter?: EntryFilter) => Promise<Result<ReadonlyArray<Entry>>>
	readonly get: (project: string, slug: string) => Promise<Result<Entry>>
	readonly getById: (project: string, id: string) => Promise<Result<Entry>>
	readonly subscribeToList: (
		project: string,
		update: (entry: Entry, action: ActionEvent) => void,
		filter?: EntryFilter,
	) => Promise<Result<Unsubscribe>>
	readonly subscribeToRecord: (
		project: string,
		id: string,
		onChange: (record: Entry) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
}

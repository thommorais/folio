import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Entry } from '_/core/domain/entry'
import { sortEntries } from '_/core/domain/order'
import type { EntryFilter } from '_/core/ports/entries'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type EntriesState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly entries: readonly Entry[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useEntries = (project: string, filter?: EntryFilter): EntriesState => {
	const { entries, connection } = useContainer()
	const key = useFilterKey(filter)

	const current = JSON.parse(key) as EntryFilter | undefined

	const state = useLiveList<Entry>({
		load: async () => {
			const result = await entries.list(project, current)
			return result.success ? { ...result, value: sortEntries(result.value, current?.sort) } : result
		},
		subscribe: update => entries.subscribeToList(project, update, current),
		fold: (rows, row, action) => sortEntries(foldUpdates(rows, row, action), current?.sort),
		connection,
		deps: [project, key, entries],
	})

	return state.status === Status.Ready ? { status: Status.Ready, entries: state.data } : state
}

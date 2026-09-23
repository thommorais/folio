import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Entry } from '_/core/domain/entry'
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

	const state = useLiveList<Entry>({
		load: () => entries.list(project, JSON.parse(key) as EntryFilter),
		subscribe: update => entries.subscribeToList(project, update, JSON.parse(key) as EntryFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, entries],
	})

	return state.status === Status.Ready ? { status: Status.Ready, entries: state.data } : state
}

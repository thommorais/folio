import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { JournalEntry } from '_/core/domain/journal'
import type { JournalFilter } from '_/core/ports/journal'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type JournalState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly journal: readonly JournalEntry[] }
	| { readonly status: 'failed'; readonly message: string }

export const useJournal = (project: string, filter?: JournalFilter): JournalState => {
	const { journal, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<JournalEntry>({
		load: () => journal.list(project, JSON.parse(key) as JournalFilter),
		subscribe: update => journal.subscribeToList(project, update, JSON.parse(key) as JournalFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, journal],
	})

	return state.status === 'ready' ? { status: 'ready', journal: state.data } : state
}

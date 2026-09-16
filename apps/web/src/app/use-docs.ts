import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Doc } from '_/core/domain/doc'
import type { DocFilter } from '_/core/ports/docs'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type DocsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly docs: readonly Doc[] }
	| { readonly status: 'failed'; readonly message: string }

export const useDocs = (project: string, filter?: DocFilter): DocsState => {
	const { docs, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Doc>({
		load: () => docs.list(project, JSON.parse(key) as DocFilter),
		subscribe: update => docs.subscribeToList(project, update, JSON.parse(key) as DocFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, docs],
	})

	return state.status === 'ready' ? { status: 'ready', docs: state.data } : state
}

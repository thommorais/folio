import type { Entry } from '_/core/domain/entry'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type EntryState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly entry: Entry }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useEntry = (project: string, slug: string): EntryState => {
	const { entries, connection } = useContainer()

	const state = useLiveRecord<Entry>({
		load: () => entries.get(project, slug),
		subscribe: (id, onChange, onGone) => entries.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === 'ready' ? { status: 'ready', entry: state.data } : state
}

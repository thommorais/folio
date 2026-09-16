import type { JournalEntry } from '_/core/domain/journal'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type JournalEntryState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly entry: JournalEntry }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useJournalEntry = (project: string, slug: string): JournalEntryState => {
	const { journal, connection } = useContainer()

	const state = useLiveRecord<JournalEntry>({
		load: () => journal.get(project, slug),
		subscribe: (id, onChange, onGone) => journal.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === 'ready' ? { status: 'ready', entry: state.data } : state
}

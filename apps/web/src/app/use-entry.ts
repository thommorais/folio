import type { Entry } from '_/core/domain/entry'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type EntryState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly entry: Entry }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useEntry = (project: string, slug: string): EntryState => {
	const { entries, connection } = useContainer()

	const state = useLiveRecord<Entry>({
		load: () => entries.get(project, slug),
		subscribe: (id, onChange, onGone) => entries.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === Status.Ready ? { status: Status.Ready, entry: state.data } : state
}

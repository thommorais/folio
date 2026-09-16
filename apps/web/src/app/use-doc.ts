import type { Doc } from '_/core/domain/doc'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type DocState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly doc: Doc }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useDoc = (project: string, slug: string): DocState => {
	const { docs, connection } = useContainer()

	const state = useLiveRecord<Doc>({
		load: () => docs.get(project, slug),
		subscribe: (id, onChange, onGone) => docs.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === 'ready' ? { status: 'ready', doc: state.data } : state
}

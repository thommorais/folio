import type { Todo } from '_/core/domain/todo'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type TodoState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly todo: Todo }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useTodo = (project: string, id: string | undefined): TodoState => {
	const { todos, connection } = useContainer()

	const state = useLiveRecord<Todo>({
		load: () => todos.get(project, id ?? ''),
		subscribe: (recordId, onChange, onGone) => todos.subscribeToRecord(project, recordId, onChange, onGone),
		connection,
		deps: [project, id],
		skip: !id,
	})

	return state.status === 'ready' ? { status: 'ready', todo: state.data } : state
}

import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Todo } from '_/core/domain/todo'
import type { TodoFilter } from '_/core/ports/todos'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type TodosState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly todos: readonly Todo[] }
	| { readonly status: 'failed'; readonly message: string }

export const useTodos = (project: string, filter?: TodoFilter): TodosState => {
	const { todos, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Todo>({
		load: () => todos.list(project, JSON.parse(key) as TodoFilter),
		subscribe: update => todos.subscribeToList(project, update, JSON.parse(key) as TodoFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, todos],
	})

	return state.status === 'ready' ? { status: 'ready', todos: state.data } : state
}

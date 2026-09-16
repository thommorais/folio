import { ActionEvent } from '_/types'

type Entity = Record<'id', unknown>

const foldUpdates = <T extends Entity>(todos: ReadonlyArray<T>, todo: T, action: ActionEvent) => {
	if (action === 'delete') {
		return todos.filter(existing => existing.id !== todo.id)
	}

	const index = todos.findIndex(existing => existing.id === todo.id)

	if (index === -1) {
		return [...todos, todo]
	}

	return todos.map((existing, at) => (at === index ? todo : existing))
}

export { foldUpdates }

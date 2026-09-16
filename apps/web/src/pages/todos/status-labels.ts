import type { TodoStatus } from '_/core/domain/todo'

export const TODO_STATUS_LABELS: Record<TodoStatus, string> = {
	pending: 'Pending',
	in_progress: 'In progress',
	done: 'Done',
	blocked: 'Blocked',
	cancelled: 'Cancelled',
}

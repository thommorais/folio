import { createFileRoute } from '@tanstack/react-router'
import { TodoDetail } from '_/pages/todo'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/todos/$todo')({
	component: TodoDetail,
})

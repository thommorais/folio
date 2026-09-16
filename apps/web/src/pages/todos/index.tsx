import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Sheet, SheetContent } from '@thom/ui/sheet'
import { useTodos } from '_/app/use-todos'
import type { Todo } from '_/core/domain/todo'
import type { TodosSearch } from '_/routes/_authenticated/$slug/todos'
import { TodoFilters } from './todo-filters'
import { TodoDetails } from './todo-details'
import { TODO_STATUS_LABELS } from './status-labels'

const Row = ({ todo, onOpen }: { readonly todo: Todo; readonly onOpen: (id: string) => void }) => (
	<li>
		<button
			type='button'
			onClick={() => {
				onOpen(todo.id)
			}}
			className='hover:bg-accent/40 flex w-full items-center gap-3 px-4 py-3 text-left transition-colors'
		>
			<span
				className={cn('border-border size-4 shrink-0 border', todo.status === 'done' && 'bg-foreground border-foreground')}
			/>

			<span className={cn('flex-1 truncate text-sm', todo.status === 'done' && 'text-dim line-through')}>
				{todo.title}
			</span>

			{todo.status === 'blocked' && <Badge color='destructive'>Blocked</Badge>}

			<span className='hidden shrink-0 items-center gap-3 sm:flex'>
				{todo.tags.map(tag => (
					<Badge key={tag} color='muted'>
						{tag}
					</Badge>
				))}
			</span>

			<span className='text-dimmer hidden w-20 shrink-0 text-right text-xs sm:block'>{todo.priority}</span>
			<span className='text-dim w-24 shrink-0 text-right text-xs'>{TODO_STATUS_LABELS[todo.status]}</span>
		</button>
	</li>
)

const Todos = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/todos' })
	const search = useSearch({ from: '/_authenticated/$slug/todos' })
	const navigate = useNavigate()

	const state = useTodos(slug, {
		ticketId: search.ticket,
		planId: search.plan,
		status: search.statuses,
		priority: search.priority,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const setTodo = (todo: string | undefined) => {
		void navigate({ from: '/$slug/todos', to: '.', search: (prev: TodosSearch) => ({ ...prev, todo }) })
	}

	return (
		<div className='space-y-4'>
			<TodoFilters />

			{state.status === 'loading' && (
				<div className='border-border divide-border divide-y border'>
					{[0, 1, 2].map(key => (
						<div key={key} className='bg-accent/40 h-11.25 animate-pulse' />
					))}
				</div>
			)}

			{state.status === 'failed' && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === 'ready' && state.todos.length === 0 && <p className='text-dim text-sm'>No todos match.</p>}

			{state.status === 'ready' && state.todos.length > 0 && (
				<ul className='border-border divide-border divide-y border'>
					{state.todos.map(todo => (
						<Row key={todo.id} todo={todo} onOpen={setTodo} />
					))}
				</ul>
			)}

			<Sheet
				open={Boolean(search.todo)}
				onOpenChange={open => {
					if (!open) {
						setTodo(undefined)
					}
				}}
			>
				<SheetContent title='Todo details'>
					<TodoDetails project={slug} todoId={search.todo} />
				</SheetContent>
			</Sheet>
		</div>
	)
}

export { Todos }

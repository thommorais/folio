import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { DropdownMenuItem } from '@thom/ui/dropdown-menu'
import { usePlans } from '_/app/use-plans'
import { useTickets } from '_/app/use-tickets'
import { PRIORITIES, TODO_STATUSES, type Priority, type TodoStatus } from '_/core/domain/todo'
import { TODO_SORT_FIELDS, type TodoSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import type { TodosSearch } from '_/routes/_authenticated/$slug/todos'
import { TODO_STATUS_LABELS } from './status-labels'
import { CONTEXT_TAGS, KIND_TAGS } from './tag-vocabulary'

const SORT_LABELS: Record<TodoSortField, string> = {
	position: 'Manual',
	title: 'Title',
	status: 'Status',
	priority: 'Priority',
	created: 'Created',
	updated: 'Updated',
}

const TodoFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/todos' })
	const search = useSearch({ from: '/_authenticated/$slug/todos' })
	const navigate = useNavigate()
	const tickets = useTickets(slug)
	const plans = usePlans(slug)

	const setFilter = (patch: Partial<TodosSearch>) => {
		void navigate({ from: '/$slug/todos', to: '.', search: (prev: TodosSearch) => ({ ...prev, ...patch }) })
	}

	const ticketTitle = (id: string): string =>
		tickets.status === 'ready' ? (tickets.tickets.find(entry => entry.id === id)?.title ?? id) : id

	const planTitle = (id: string): string =>
		plans.status === 'ready' ? (plans.plans.find(entry => entry.id === id)?.title ?? id) : id

	const chips: ActiveFilter[] = []

	if (search.ticket !== undefined) {
		chips.push({
			key: 'ticket',
			label: ticketTitle(search.ticket),
			onRemove: () => {
				setFilter({ ticket: undefined })
			},
		})
	}
	if (search.plan !== undefined) {
		chips.push({
			key: 'plan',
			label: planTitle(search.plan),
			onRemove: () => {
				setFilter({ plan: undefined })
			},
		})
	}
	if (search.statuses !== undefined) {
		chips.push({
			key: 'statuses',
			label: search.statuses.map(status => TODO_STATUS_LABELS[status]).join(', '),
			onRemove: () => {
				setFilter({ statuses: undefined })
			},
		})
	}
	if (search.priority !== undefined) {
		chips.push({
			key: 'priority',
			label: search.priority,
			onRemove: () => {
				setFilter({ priority: undefined })
			},
		})
	}
	if (search.tags !== undefined) {
		chips.push({
			key: 'tags',
			label: search.tags.join(', '),
			onRemove: () => {
				setFilter({ tags: undefined })
			},
		})
	}

	return (
		<FilterBar
			placeholder='Search todos...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={TODO_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Status'>
				{TODO_STATUSES.map(status => (
					<FilterCheckboxItem
						key={status}
						label={TODO_STATUS_LABELS[status]}
						checked={search.statuses?.includes(status) ?? false}
						onCheckedChange={() => {
							setFilter({ statuses: toggle<TodoStatus>(search.statuses, status) })
						}}
					/>
				))}
			</FilterMenuItem>

			<FilterMenuItem label='Priority'>
				{PRIORITIES.map(priority => (
					<FilterCheckboxItem
						key={priority}
						label={priority}
						checked={search.priority === priority}
						onCheckedChange={() => {
							setFilter({ priority: search.priority === priority ? undefined : (priority as Priority) })
						}}
					/>
				))}
			</FilterMenuItem>

			<FilterMenuItem label='Ticket'>
				<div className='max-h-[300px] overflow-y-auto'>
					{tickets.status === 'ready' && tickets.tickets.length === 0 && (
						<DropdownMenuItem disabled>No tickets found</DropdownMenuItem>
					)}
					{tickets.status === 'ready' &&
						tickets.tickets.map(ticket => (
							<FilterCheckboxItem
								key={ticket.id}
								label={ticket.title}
								checked={search.ticket === ticket.id}
								onCheckedChange={() => {
									setFilter({ ticket: search.ticket === ticket.id ? undefined : ticket.id })
								}}
							/>
						))}
				</div>
			</FilterMenuItem>

			<FilterMenuItem label='Plan'>
				<div className='max-h-[300px] overflow-y-auto'>
					{plans.status === 'ready' && plans.plans.length === 0 && (
						<DropdownMenuItem disabled>No plans found</DropdownMenuItem>
					)}
					{plans.status === 'ready' &&
						plans.plans.map(plan => (
							<FilterCheckboxItem
								key={plan.id}
								label={plan.title}
								checked={search.plan === plan.id}
								onCheckedChange={() => {
									setFilter({ plan: search.plan === plan.id ? undefined : plan.id })
								}}
							/>
						))}
				</div>
			</FilterMenuItem>

			<FilterMenuItem label='Tags'>
				<div className='max-h-[300px] overflow-y-auto'>
					{[...CONTEXT_TAGS, ...KIND_TAGS].map(tag => (
						<FilterCheckboxItem
							key={tag}
							label={tag}
							checked={search.tags?.includes(tag) ?? false}
							onCheckedChange={() => {
								setFilter({ tags: toggle(search.tags, tag) })
							}}
						/>
					))}
				</div>
			</FilterMenuItem>
		</FilterBar>
	)
}

export { TodoFilters }

import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { DropdownMenuItem } from '@thom/ui/dropdown-menu'
import { useTickets } from '_/app/use-tickets'
import { PLAN_STATUSES, type PlanStatus } from '_/core/domain/plan'
import { PLAN_SORT_FIELDS, type PlanSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/todos/tag-vocabulary'
import type { PlansSearch } from '_/routes/_authenticated/$slug/plans/index'
import { PLAN_STATUS_LABELS } from './status-labels'

const SORT_LABELS: Record<PlanSortField, string> = {
	title: 'Title',
	status: 'Status',
	created: 'Created',
	updated: 'Updated',
}

const PlanFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/plans/' })
	const search = useSearch({ from: '/_authenticated/$slug/plans/' })
	const navigate = useNavigate()
	const tickets = useTickets(slug)

	const setFilter = (patch: Partial<PlansSearch>) => {
		void navigate({ from: '/$slug/plans/', to: '.', search: (prev: PlansSearch) => ({ ...prev, ...patch }) })
	}

	const ticketTitle = (id: string): string =>
		tickets.status === 'ready' ? (tickets.tickets.find(entry => entry.id === id)?.title ?? id) : id

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
	if (search.statuses !== undefined) {
		chips.push({
			key: 'statuses',
			label: search.statuses.map(status => PLAN_STATUS_LABELS[status]).join(', '),
			onRemove: () => {
				setFilter({ statuses: undefined })
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
			placeholder='Search plans...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={PLAN_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Status'>
				{PLAN_STATUSES.map(status => (
					<FilterCheckboxItem
						key={status}
						label={PLAN_STATUS_LABELS[status]}
						checked={search.statuses?.includes(status) ?? false}
						onCheckedChange={() => {
							setFilter({ statuses: toggle<PlanStatus>(search.statuses, status) })
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

export { PlanFilters }

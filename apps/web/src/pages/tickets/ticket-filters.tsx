import { useNavigate, useSearch } from '@tanstack/react-router'
import { TICKET_STATUSES, type TicketStatus } from '_/core/domain/ticket'
import { PRIORITIES, type Priority } from '_/core/domain/todo'
import { TICKET_SORT_FIELDS, type TicketSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/todos/tag-vocabulary'
import type { TicketsSearch } from '_/routes/_authenticated/$slug/tickets'
import { TICKET_STATUS_LABELS } from './status-labels'

const SORT_LABELS: Record<TicketSortField, string> = {
	title: 'Title',
	status: 'Status',
	priority: 'Priority',
	created: 'Created',
	updated: 'Updated',
}

const TicketFilters = () => {
	const search = useSearch({ from: '/_authenticated/$slug/tickets/' })
	const navigate = useNavigate()

	const setFilter = (patch: Partial<TicketsSearch>) => {
		void navigate({ from: '/$slug/tickets/', to: '.', search: (prev: TicketsSearch) => ({ ...prev, ...patch }) })
	}

	const chips: ActiveFilter[] = []

	if (search.statuses !== undefined) {
		chips.push({
			key: 'statuses',
			label: search.statuses.map(status => TICKET_STATUS_LABELS[status]).join(', '),
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
			placeholder='Search tickets...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={TICKET_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Status'>
				{TICKET_STATUSES.map(status => (
					<FilterCheckboxItem
						key={status}
						label={TICKET_STATUS_LABELS[status]}
						checked={search.statuses?.includes(status) ?? false}
						onCheckedChange={() => {
							setFilter({ statuses: toggle<TicketStatus>(search.statuses, status) })
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

export { TicketFilters }

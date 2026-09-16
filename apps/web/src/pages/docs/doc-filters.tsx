import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { DropdownMenuItem } from '@thom/ui/dropdown-menu'
import { useTickets } from '_/app/use-tickets'
import { DOC_SORT_FIELDS, type DocSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/todos/tag-vocabulary'
import type { DocsSearch } from '_/routes/_authenticated/$slug/docs'

const SORT_LABELS: Record<DocSortField, string> = {
	title: 'Title',
	slug: 'Slug',
	created: 'Created',
	updated: 'Updated',
}

const DocsFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/docs/' })
	const search = useSearch({ from: '/_authenticated/$slug/docs/' })
	const navigate = useNavigate()
	const tickets = useTickets(slug)

	const setFilter = (patch: Partial<DocsSearch>) => {
		void navigate({ from: '/$slug/docs/', to: '.', search: (prev: DocsSearch) => ({ ...prev, ...patch }) })
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
			placeholder='Search docs...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={DOC_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
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

export { DocsFilters }

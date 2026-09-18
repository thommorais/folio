import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { DropdownMenuItem } from '@thom/ui/dropdown-menu'
import { useIssues } from '_/app/use-issues'
import { ENTRY_SORT_FIELDS, type EntrySortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary'
import type { LogsSearch } from '_/routes/_authenticated/$client/$domain/$slug/journal'

const SORT_LABELS: Record<EntrySortField, string> = {
	title: 'Title',
	slug: 'Slug',
	created: 'Created',
	updated: 'Updated',
}

const LogsFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const navigate = useNavigate()
	const tickets = useIssues(slug, { kind: 'ticket' })

	const setFilter = (patch: Partial<LogsSearch>) => {
		void navigate({
			from: '/$client/$domain/$slug/journal/',
			to: '.',
			search: (prev: LogsSearch) => ({ ...prev, ...patch }),
		})
	}

	const ticketTitle = (id: string): string =>
		tickets.status === 'ready' ? (tickets.issues.find(entry => entry.id === id)?.title ?? id) : id

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
			placeholder='Search logs...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={ENTRY_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Issue'>
				<div className='max-h-[300px] overflow-y-auto'>
					{tickets.status === 'ready' && tickets.issues.length === 0 && (
						<DropdownMenuItem disabled>No tickets found</DropdownMenuItem>
					)}
					{tickets.status === 'ready' &&
						tickets.issues.map(ticket => (
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

export { LogsFilters }

import { useNavigate, useParams, useSearch } from '@tanstack/react-router';
import { DropdownMenuItem } from '@thom/ui/dropdown-menu';
import { useIssues } from '_/app/use-issues';
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle, FILTER_KEY } from '_/components/list/filter-bar';
import { SortMenu } from '_/components/list/sort-menu';
import { ADDRESSABLE_KINDS, type AddressableKind } from '_/core/domain/entry';
import { ENTRY_SORT_FIELDS, type EntrySortField } from '_/core/ports/sort';
import { Status } from '_/lib/async-status';
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary';
import type { JournalSearch } from '_/routes/_authenticated/$client/$domain/$slug/journal';
import { ENTRY_KIND_LABELS } from './kind-labels';
import { ISSUE_KIND } from '_/core/domain/issue';

const SORT_LABELS: Record<EntrySortField, string> = {
	title: 'Title',
	slug: 'Slug',
	created: 'Created',
	updated: 'Updated',
}

const JournalFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const navigate = useNavigate()
	const tickets = useIssues(slug, { kind: ISSUE_KIND.TICKET })

	const setFilter = (patch: Partial<JournalSearch>) => {
		void navigate({
			from: '/$client/$domain/$slug/journal/',
			to: '.',
			search: (prev: JournalSearch) => ({ ...prev, ...patch }),
		})
	}

	const ticketTitle = (id: string): string =>
		tickets.status === Status.Ready ? (tickets.issues.find(entry => entry.id === id)?.title ?? id) : id

	const chips: ActiveFilter[] = []

	if (search.kinds !== undefined) {
		chips.push({
			key: FILTER_KEY.KINDS,
			label: search.kinds.map(kind => ENTRY_KIND_LABELS[kind]).join(', '),
			onRemove: () => {
				setFilter({ kinds: undefined })
			},
		})
	}
	if (search.ticket !== undefined) {
		chips.push({
			key: FILTER_KEY.TICKET,
			label: ticketTitle(search.ticket),
			onRemove: () => {
				setFilter({ ticket: undefined })
			},
		})
	}
	if (search.tags !== undefined) {
		chips.push({
			key: FILTER_KEY.TAGS,
			label: search.tags.join(', '),
			onRemove: () => {
				setFilter({ tags: undefined })
			},
		})
	}

	return (
		<FilterBar
			placeholder='Search the journal...'
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
			<FilterMenuItem label='Type'>
				{ADDRESSABLE_KINDS.map(kind => (
					<FilterCheckboxItem
						key={kind}
						label={ENTRY_KIND_LABELS[kind]}
						checked={search.kinds?.includes(kind) ?? false}
						onCheckedChange={() => {
							setFilter({ kinds: toggle<AddressableKind>(search.kinds, kind) })
						}}
					/>
				))}
			</FilterMenuItem>

			<FilterMenuItem label='Ticket'>
				<div className='max-h-75 overflow-y-auto'>
					{tickets.status === Status.Ready && tickets.issues.length === 0 && (
						<DropdownMenuItem disabled>No tickets found</DropdownMenuItem>
					)}
					{tickets.status === Status.Ready &&
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
				<div className='max-h-75 overflow-y-auto'>
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

export { JournalFilters };

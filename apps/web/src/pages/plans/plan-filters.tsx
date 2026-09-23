import { useNavigate, useParams, useSearch } from '@tanstack/react-router';
import { DropdownMenuItem } from '@thom/ui/dropdown-menu';
import { useIssues } from '_/app/use-issues';
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle, FILTER_KEY } from '_/components/list/filter-bar';
import { SortMenu } from '_/components/list/sort-menu';
import { DEFAULT_PLAN_STATUSES, PLAN_STATUSES, type PlanStatus } from '_/core/domain/plan';
import { PLAN_SORT_FIELDS, type PlanSortField } from '_/core/ports/sort';
import { Status } from '_/lib/async-status';
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary';
import type { PlansSearch } from '_/routes/_authenticated/$client/$domain/$slug/plans/index';
import { PLAN_STATUS_LABELS } from './status-labels';

const SORT_LABELS: Record<PlanSortField, string> = {
	title: 'Title',
	status: 'Status',
	created: 'Created',
	updated: 'Updated',
}

const PlanFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/plans/' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/plans/' })
	const navigate = useNavigate()
	const tickets = useIssues(slug)
	const statuses = search.statuses ?? DEFAULT_PLAN_STATUSES

	const setFilter = (patch: Partial<PlansSearch>) => {
		void navigate({
			from: '/$client/$domain/$slug/plans/',
			to: '.',
			search: (prev: PlansSearch) => ({ ...prev, ...patch }),
		})
	}

	const ticketTitle = (id: string): string =>
		tickets.status === Status.Ready ? (tickets.issues.find(entry => entry.id === id)?.title ?? id) : id

	const chips: ActiveFilter[] = []

	if (search.ticket !== undefined) {
		chips.push({
			key: FILTER_KEY.TICKET,
			label: ticketTitle(search.ticket),
			onRemove: () => {
				setFilter({ ticket: undefined })
			},
		})
	}
	if (search.statuses !== undefined) {
		chips.push({
			key: FILTER_KEY.STATUSES,
			label: search.statuses.map(status => PLAN_STATUS_LABELS[status]).join(', '),
			onRemove: () => {
				setFilter({ statuses: undefined })
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
						checked={statuses.includes(status)}
						onCheckedChange={() => {
							setFilter({ statuses: toggle<PlanStatus>(statuses, status) })
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

export { PlanFilters };

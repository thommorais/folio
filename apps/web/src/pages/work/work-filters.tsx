import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { DropdownMenuItem } from '@thom/ui/dropdown-menu'
import { useIssues } from '_/app/use-issues'
import { ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue'
import { PLAN_STATUSES, type PlanStatus } from '_/core/domain/plan'
import { WORK_SORT_FIELDS, type WorkSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary'
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels'
import type { WorkSearch } from '_/routes/_authenticated/$client/$domain/$slug/work'
import { WORK_TYPES, WORK_TYPE_LABELS, type WorkType } from './types'
import { Status } from '_/lib/async-status'

const SORT_LABELS: Record<WorkSortField, string> = {
	title: 'Title',
	status: 'Status',
	created: 'Created',
	updated: 'Updated',
}

const WorkFilters = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/work' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/work' })
	const navigate = useNavigate()
	const tickets = useIssues(slug, { kind: 'ticket' })

	const setFilter = (patch: Partial<WorkSearch>) => {
		void navigate({
			from: '/$client/$domain/$slug/work',
			to: '.',
			search: (prev: WorkSearch) => ({ ...prev, ...patch }),
		})
	}

	const ticketTitle = (id: string): string =>
		tickets.status === Status.Ready ? (tickets.issues.find(entry => entry.id === id)?.title ?? id) : id

	const chips: ActiveFilter[] = []

	if (search.types !== undefined) {
		chips.push({
			key: 'types',
			label: search.types.map(type => WORK_TYPE_LABELS[type]).join(', '),
			onRemove: () => {
				setFilter({ types: undefined })
			},
		})
	}
	if (search.statuses !== undefined) {
		chips.push({
			key: 'statuses',
			label: search.statuses.map(status => ISSUE_STATUS_LABELS[status]).join(', '),
			onRemove: () => {
				setFilter({ statuses: undefined })
			},
		})
	}
	if (search.planStatuses !== undefined) {
		chips.push({
			key: 'planStatuses',
			label: search.planStatuses.map(status => PLAN_STATUS_LABELS[status]).join(', '),
			onRemove: () => {
				setFilter({ planStatuses: undefined })
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
	if (search.ticket !== undefined) {
		chips.push({
			key: 'ticket',
			label: ticketTitle(search.ticket),
			onRemove: () => {
				setFilter({ ticket: undefined })
			},
		})
	}

	return (
		<FilterBar
			placeholder='Search tickets, plans and todos...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={WORK_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Type'>
				{WORK_TYPES.map(type => (
					<FilterCheckboxItem
						key={type}
						label={WORK_TYPE_LABELS[type]}
						checked={search.types?.includes(type) ?? false}
						onCheckedChange={() => {
							setFilter({ types: toggle<WorkType>(search.types, type) })
						}}
					/>
				))}
			</FilterMenuItem>

			<FilterMenuItem label='Ticket & todo status'>
				{ISSUE_STATUSES.map(status => (
					<FilterCheckboxItem
						key={status}
						label={ISSUE_STATUS_LABELS[status]}
						checked={search.statuses?.includes(status) ?? false}
						onCheckedChange={() => {
							setFilter({ statuses: toggle<IssueStatus>(search.statuses, status) })
						}}
					/>
				))}
			</FilterMenuItem>

			<FilterMenuItem label='Plan status'>
				{PLAN_STATUSES.map(status => (
					<FilterCheckboxItem
						key={status}
						label={PLAN_STATUS_LABELS[status]}
						checked={search.planStatuses?.includes(status) ?? false}
						onCheckedChange={() => {
							setFilter({ planStatuses: toggle<PlanStatus>(search.planStatuses, status) })
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

export { WorkFilters }

import { useNavigate, useSearch } from '@tanstack/react-router'
import { ISSUE_STATUSES, type IssueStatus } from '_/core/domain/issue'
import { PRIORITIES, type Priority } from '_/core/domain/issue'
import { ISSUE_SORT_FIELDS, type IssueSortField } from '_/core/ports/sort'
import { ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary'
import { ISSUE_STATUS_LABELS } from './status-labels'

const SORT_LABELS: Record<IssueSortField, string> = {
	title: 'Title',
	status: 'Status',
	priority: 'Priority',
	size: 'Size',
	position: 'Position',
	created: 'Created',
	updated: 'Updated',
}

type IssuesSearch = {
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: { readonly field: IssueSortField; readonly direction: 'asc' | 'desc' }
}

const IssueFilters = () => {
	const search = useSearch({ strict: false }) as IssuesSearch
	const navigate = useNavigate()

	const setFilter = (patch: Partial<IssuesSearch>) => {
		void navigate({ to: '.', search: (prev: Record<string, unknown>) => ({ ...prev, ...patch }) })
	}

	const chips: ActiveFilter[] = []

	if (search.statuses !== undefined) {
		chips.push({
			key: 'statuses',
			label: search.statuses.map(status => ISSUE_STATUS_LABELS[status]).join(', '),
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
			placeholder='Search issues...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={ISSUE_SORT_FIELDS}
					labels={SORT_LABELS}
					sort={search.sort}
					onChange={sort => {
						setFilter({ sort })
					}}
				/>
			}
		>
			<FilterMenuItem label='Status'>
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

export { IssueFilters }

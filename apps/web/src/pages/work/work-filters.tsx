import { useNavigate, useSearch } from '@tanstack/react-router'
import { DEFAULT_ISSUE_STATUSES, ISSUE_STATUSES, PRIORITIES, type IssueStatus, type Priority } from '_/core/domain/issue'
import { defaultChip } from '_/components/list/default-chip'
import { WORK_SORT_FIELDS, type WorkSortField } from '_/core/ports/sort'
import { chipPerValue, ActiveFilter, FilterBar, FilterCheckboxItem, FilterMenuItem, toggle, FILTER_KEY } from '_/components/list/filter-bar'
import { SortMenu } from '_/components/list/sort-menu'
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels'
import { CONTEXT_TAGS, KIND_TAGS } from '_/pages/issues/tag-vocabulary'
import type { WorkSearch } from '_/routes/_authenticated/$client/$domain/$slug/work'

const SORT_LABELS: Record<WorkSortField, string> = {
	title: 'Title',
	status: 'Status',
	created: 'Created',
	updated: 'Updated',
}

const WorkFilters = () => {
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/work' })
	const navigate = useNavigate()

	const setFilter = (patch: Partial<WorkSearch>) => {
		void navigate({
			from: '/$client/$domain/$slug/work',
			to: '.',
			search: (prev: WorkSearch) => ({ ...prev, ...patch }),
		})
	}

	const chips: ActiveFilter[] = []

	const hidingIssues = defaultChip(FILTER_KEY.STATUSES, search.statuses, ISSUE_STATUSES, DEFAULT_ISSUE_STATUSES, status => ISSUE_STATUS_LABELS[status].toLowerCase(), all => {
		setFilter({ statuses: all })
	})
	if (hidingIssues) chips.push(hidingIssues)

	chips.push(
		...chipPerValue(FILTER_KEY.STATUSES, search.statuses, status => ISSUE_STATUS_LABELS[status], statuses => {
			setFilter({ statuses })
		}),
	)
	if (search.priority !== undefined) {
		chips.push({
			key: FILTER_KEY.PRIORITY,
			label: search.priority,
			onRemove: () => {
				setFilter({ priority: undefined })
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
			placeholder='Search tickets...'
			term={search.q}
			onSearch={q => {
				setFilter({ q })
			}}
			chips={chips}
			trailing={
				<SortMenu
					fields={WORK_SORT_FIELDS}
					labels={SORT_LABELS}
					defaultLabel='Default order'
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

export { WorkFilters }

import { useParams, useSearch } from '@tanstack/react-router'
import { useIssues } from '_/app/use-issues'
import { usePlans } from '_/app/use-plans'
import { isTerminal, type Issue } from '_/core/domain/issue'
import { isTerminal as isPlanTerminal, type Plan } from '_/core/domain/plan'
import { useScope } from '_/routing/use-scope'
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels'
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels'
import { Section, type WorkItem } from './section'
import type { WorkType } from './types'
import { WorkFilters } from './work-filters'

type Scope = {
	readonly client: string
	readonly domain: string
	readonly slug: string
}

const issueItem = (scope: Scope, issue: Issue): WorkItem => ({
	id: issue.id,
	title: issue.title,
	status: ISSUE_STATUS_LABELS[issue.status],
	muted: isTerminal(issue.status),
	meta: [issue.priority, ...issue.tags],
	to: '/$client/$domain/$slug/tickets/$ticket',
	params: { ...scope, ticket: issue.slug },
})

// A todo has no page of its own: the todos list opens it in the preview sheet,
// so the link carries the id as a search param there instead.
const todoItem = (scope: Scope, todo: Issue): WorkItem => ({
	id: todo.id,
	title: todo.title,
	status: ISSUE_STATUS_LABELS[todo.status],
	muted: isTerminal(todo.status),
	checked: todo.status === 'done',
	meta: [todo.priority, ...todo.tags],
	to: '/$client/$domain/$slug/todos',
	params: scope,
})

const planItem = (scope: Scope, plan: Plan): WorkItem => ({
	id: plan.id,
	title: plan.title,
	status: PLAN_STATUS_LABELS[plan.status],
	muted: isPlanTerminal(plan.status),
	meta: plan.tags,
	to: '/$client/$domain/$slug/plans/$plan',
	params: { ...scope, plan: plan.id },
})

const Work = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/work' })
	const { client, domain } = useScope()
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/work' })

	const scope: Scope = { client, domain, slug }
	const term = search.q
	// No selection means every section shows, which is what an untouched filter
	// should do.
	const shows = (type: WorkType): boolean => search.types === undefined || search.types.includes(type)

	const filtered =
		term !== undefined ||
		search.types !== undefined ||
		search.statuses !== undefined ||
		search.planStatuses !== undefined ||
		search.priority !== undefined ||
		search.tags !== undefined ||
		search.ticket !== undefined

	const tickets = useIssues(slug, {
		kind: 'ticket',
		search: term,
		status: search.statuses,
		priority: search.priority,
		tags: search.tags,
		sort: search.sort,
	})
	const todos = useIssues(slug, {
		kind: 'todo',
		search: term,
		status: search.statuses,
		priority: search.priority,
		tags: search.tags,
		parentId: search.ticket,
		sort: search.sort,
	})
	// Plans carry no priority, so that filter narrows the other two sections only.
	const plans = usePlans(slug, {
		search: term,
		status: search.planStatuses,
		tags: search.tags,
		ticketId: search.ticket,
		sort: search.sort,
	})

	return (
		<div className='space-y-4'>
			<WorkFilters />

			<div className='space-y-8'>
				{shows('tickets') && (
					<Section
						title='Tickets'
						items={tickets.status === 'ready' ? tickets.issues.map(issue => issueItem(scope, issue)) : undefined}
						message={tickets.status === 'failed' ? tickets.message : undefined}
						emptyLabel='tickets'
						filtered={filtered}
						to='/$client/$domain/$slug/tickets'
						params={scope}
					/>
				)}

				{shows('plans') && (
					<Section
						title='Plans'
						items={plans.status === 'ready' ? plans.plans.map(plan => planItem(scope, plan)) : undefined}
						message={plans.status === 'failed' ? plans.message : undefined}
						emptyLabel='plans'
						filtered={filtered}
						to='/$client/$domain/$slug/plans'
						params={scope}
					/>
				)}

				{shows('todos') && (
					<Section
						title='Todos'
						items={todos.status === 'ready' ? todos.issues.map(todo => todoItem(scope, todo)) : undefined}
						message={todos.status === 'failed' ? todos.message : undefined}
						emptyLabel='todos'
						filtered={filtered}
						to='/$client/$domain/$slug/todos'
						params={scope}
					/>
				)}
			</div>
		</div>
	)
}

export { Work }

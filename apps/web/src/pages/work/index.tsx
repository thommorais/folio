import { useParams, useSearch } from '@tanstack/react-router';
import { useIssues } from '_/app/use-issues';
import { usePlans } from '_/app/use-plans';
import { ISSUE_KIND, ISSUE_STATUS, isTerminal, type Issue } from '_/core/domain/issue';
import { isTerminal as isPlanTerminal, PLAN_STATUS, type Plan } from '_/core/domain/plan';
import { Status } from '_/lib/async-status';
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels';
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels';
import { useScope } from '_/routing/use-scope';
import { Section, type WorkItem } from './section';
import type { WorkType } from './types';
import { WorkFilters } from './work-filters';

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

const todoItem = (scope: Scope, todo: Issue): WorkItem => ({
	id: todo.id,
	title: todo.title,
	status: ISSUE_STATUS_LABELS[todo.status],
	muted: isTerminal(todo.status),
	checked: todo.status === ISSUE_STATUS.DONE,
	meta: [todo.priority, ...todo.tags],
	to: '/$client/$domain/$slug/tickets/$ticket',
	params: { ...scope, ticket: todo.slug },
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
		kind: ISSUE_KIND.TICKET,
		search: term,
		status: search.statuses || [ISSUE_STATUS.IN_PROGRESS, ISSUE_STATUS.OPEN, ISSUE_STATUS.BLOCKED],
		priority: search.priority,
		tags: search.tags,
		sort: search.sort,
	})

	const todos = useIssues(slug, {
		kind: ISSUE_KIND.TODO,
		search: term,
		status: search.statuses || [ISSUE_STATUS.IN_PROGRESS, ISSUE_STATUS.OPEN, ISSUE_STATUS.BLOCKED],
		priority: search.priority,
		tags: search.tags,
		parentId: search.ticket,
		sort: search.sort,
	})

	const plans = usePlans(slug, {
		search: term,
		status: search.planStatuses || [PLAN_STATUS.ACTIVE],
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
						items={tickets.status === Status.Ready ? tickets.issues.map(issue => issueItem(scope, issue)) : undefined}
						message={tickets.status === Status.Failed ? tickets.message : undefined}
						emptyLabel='tickets'
						filtered={filtered}
						to='/$client/$domain/$slug/tickets'
						params={scope}
					/>
				)}

				{shows('plans') && (
					<Section
						title='Plans'
						items={plans.status === Status.Ready ? plans.plans.map(plan => planItem(scope, plan)) : undefined}
						message={plans.status === Status.Failed ? plans.message : undefined}
						emptyLabel='plans'
						filtered={filtered}
						to='https://folio.journ.app/welligence/web/xwwp/tickets/xwwp-5227-update-apollo-js/$client/$domain/$slug/plans'
						params={scope}
					/>
				)}

				{shows('todos') && (
					<Section
						title='Todos'
						items={todos.status === Status.Ready ? todos.issues.map(todo => todoItem(scope, todo)) : undefined}
						message={todos.status === Status.Failed ? todos.message : undefined}
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

export { Work };

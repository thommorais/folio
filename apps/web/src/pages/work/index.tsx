import { useParams, useSearch } from '@tanstack/react-router';
import { useIssues } from '_/app/use-issues';
import { usePlans } from '_/app/use-plans';
import { cn } from '@thom/libs/cn';
import { IssueLine } from '_/components/issue/issue-line';
import { Tags } from '_/components/issue/tag';
import { DEFAULT_ISSUE_STATUSES, ISSUE_KIND, type Issue } from '_/core/domain/issue';
import { isTopLevel } from '_/core/domain/top-level';
import { DEFAULT_PLAN_STATUSES, isTerminal as isPlanTerminal, type Plan } from '_/core/domain/plan';
import { Status } from '_/lib/async-status';
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
	to: '/$client/$domain/$slug/tickets/$ticket',
	params: { ...scope, ticket: issue.slug },
	line: <IssueLine issue={issue} />,
	changedAt: issue.updatedAt.getTime(),
})

const todoItem = (scope: Scope, todo: Issue): WorkItem => ({
	id: todo.id,
	to: '/$client/$domain/$slug/todos/$todo',
	params: { ...scope, todo: todo.slug },
	line: <IssueLine issue={todo} strike />,
	changedAt: todo.updatedAt.getTime(),
})

const planItem = (scope: Scope, plan: Plan): WorkItem => ({
	id: plan.id,
	to: '/$client/$domain/$slug/plans/$plan',
	params: { ...scope, plan: plan.id },
	changedAt: plan.updatedAt.getTime(),
	line: (
		<div className='flex h-5 items-center gap-3'>
			<span className={cn('min-w-0 flex-1 truncate text-sm', isPlanTerminal(plan.status) && 'text-dim')}>
				{plan.title}
			</span>
			<Tags tags={plan.tags} className='hidden shrink-0 flex-nowrap sm:flex' />
			<span className='text-dim w-20 shrink-0 text-right text-xs'>{PLAN_STATUS_LABELS[plan.status]}</span>
		</div>
	),
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
		status: search.statuses ?? DEFAULT_ISSUE_STATUSES,
		priority: search.priority,
		tags: search.tags,
		sort: search.sort,
	})

	const todos = useIssues(slug, {
		kind: ISSUE_KIND.TODO,
		search: term,
		status: search.statuses ?? DEFAULT_ISSUE_STATUSES,
		priority: search.priority,
		tags: search.tags,
		parentId: search.ticket,
		sort: search.sort,
	})

	const topLevel = (issues: readonly Issue[]): readonly Issue[] =>
		search.ticket === undefined ? issues.filter(isTopLevel) : issues

	const plans = usePlans(slug, {
		search: term,
		status: search.planStatuses ?? DEFAULT_PLAN_STATUSES,
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
						items={tickets.status === Status.Ready ? topLevel(tickets.issues).map(issue => issueItem(scope, issue)) : undefined}
						message={tickets.status === Status.Failed ? tickets.message : undefined}
						emptyLabel='tickets'
						command={`folio ticket create "<title>" -p ${slug}`}
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
						command={`folio plan create "<title>" -p ${slug}`}
						filtered={filtered}
						to='https://folio.journ.app/welligence/web/xwwp/tickets/xwwp-5227-update-apollo-js/$client/$domain/$slug/plans'
						params={scope}
					/>
				)}

				{shows('todos') && (
					<Section
						title='Todos'
						items={todos.status === Status.Ready ? topLevel(todos.issues).map(todo => todoItem(scope, todo)) : undefined}
						message={todos.status === Status.Failed ? todos.message : undefined}
						emptyLabel='todos'
						command={`folio todo create "<title>" -p ${slug}`}
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

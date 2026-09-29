import { useParams, useSearch } from '@tanstack/react-router';
import { useIssues } from '_/app/use-issues';
import { IssueLine } from '_/components/issue/issue-line';
import { DEFAULT_ISSUE_STATUSES, ISSUE_KIND, type Issue } from '_/core/domain/issue';
import { isTopLevel } from '_/core/domain/top-level';
import { Status } from '_/lib/async-status';
import { useScope } from '_/routing/use-scope';
import { Section, type WorkItem } from './section';
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

const Work = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/work' })
	const { client, domain } = useScope()
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/work' })

	const scope: Scope = { client, domain, slug }
	const term = search.q

	const filtered =
		term !== undefined ||
		search.statuses !== undefined ||
		search.priority !== undefined ||
		search.tags !== undefined

	const tickets = useIssues(slug, {
		kind: ISSUE_KIND.TICKET,
		search: term,
		status: search.statuses ?? DEFAULT_ISSUE_STATUSES,
		priority: search.priority,
		tags: search.tags,
		sort: search.sort,
	})

	return (
		<div className='space-y-4'>
			<WorkFilters />

			<Section
				title='Tickets'
				items={tickets.status === Status.Ready ? tickets.issues.filter(isTopLevel).map(issue => issueItem(scope, issue)) : undefined}
				message={tickets.status === Status.Failed ? tickets.message : undefined}
				emptyLabel='tickets'
				command={`folio ticket create "<title>" -p ${slug}`}
				filtered={filtered}
				to='/$client/$domain/$slug/tickets'
				params={scope}
			/>
		</div>
	)
}

export { Work };

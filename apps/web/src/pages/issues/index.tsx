import { Link, useParams, useSearch } from '@tanstack/react-router';
import { cn } from '@thom/libs/cn';
import { Badge } from '@thom/ui/badge';
import { useIssues } from '_/app/use-issues';
import { usePlans } from '_/app/use-plans';
import { GroupHeader } from '_/components/group/group-header';
import { groupTodos } from '_/core/domain/grouping';
import { MarkdownPreview } from '_/components/markdown/preview';
import { ISSUE_LINE_INSET, IssueLine } from '_/components/issue/issue-line';
import { StaggerItem } from '_/components/motion/stagger';
import { Flash } from '_/components/motion/flash';
import { isTerminal, ISSUE_KIND, type Issue, type IssueKind, type IssueStatus, type Priority } from '_/core/domain/issue';
import { buildIssueTree, type IssueRow } from '_/core/domain/issue-tree';
import type { IssueSortField, Sort } from '_/core/ports/sort';
import { Status } from '_/lib/async-status';
import { useScope } from '_/routing/use-scope';
import { EmptyState } from '_/components/empty-state';
import { IssueFilters } from './issue-filters';
import { LoadError } from '_/components/load-error';
import { ExternalRef } from '_/components/issue/external-ref';

type IssuesSearch = {
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<IssueSortField>
}

const INDENT = 22
// Vertical center of a nested row's title line, where the connector meets it.
const ELBOW = '1.25rem'

// Guides live in the row's left gutter rather than in the flow, so the title
// column keeps one offset per depth instead of drifting with the markup.
const Guides = ({ row }: { readonly row: IssueRow }) => {
	if (row.depth === 0) return null

	const trunk = (row.depth - 1) * INDENT + INDENT / 2

	return (
		// Guides run a pixel past the row on each side so the divider border
		// between rows does not read as a break in the line.
		<span aria-hidden className='pointer-events-none absolute -top-px -bottom-px left-0 w-0'>
			{row.guides.map((continues, level) =>
				continues ? (
					<span
						// Guides are positional; the level is the only identity they have.
						key={level}
						className='border-border absolute inset-y-0 border-l'
						style={{ left: level * INDENT + INDENT / 2 }}
					/>
				) : null,
			)}

			<span
				className='border-border absolute top-0 border-l'
				style={{ left: trunk, height: row.isLast ? ELBOW : '100%' }}
			/>

			<span className='border-border absolute w-2 border-t' style={{ left: trunk, top: ELBOW }} />
		</span>
	)
}

const Row = ({
	row,
	project,
	kind,
}: {
	readonly row: IssueRow
	readonly project: string
	readonly kind: IssueKind
}) => {
	const { issue } = row
	const { client, domain } = useScope()
	const muted = isTerminal(issue.status) || row.isContext

	return (
		<article className='hover:bg-accent/40 active:bg-accent/60 has-focus-visible:bg-accent/40 relative px-4 py-2.5 transition-colors'>
			<Flash on={issue.updatedAt.getTime()} />
			<Guides row={row} />

			<div style={{ paddingLeft: row.depth * INDENT }}>
				<IssueLine
					issue={row.isContext ? { ...issue, tags: [] } : issue}
					muted={muted}
					strike={kind === ISSUE_KIND.TODO}
					title={
						<Link
							to={kind === ISSUE_KIND.TODO ? '/$client/$domain/$slug/todos/$todo' : '/$client/$domain/$slug/tickets/$ticket'}
							params={{ client, domain, slug: project, ticket: issue.slug, todo: issue.slug }}
							data-row
							className='after:absolute after:inset-0'
						>
							{issue.title}
						</Link>
					}
					extra={
						!row.isContext && (
							<>
								{issue.wayfinder && <Badge color='muted'>{issue.wayfinder}</Badge>}
								{issue.externalRef && <ExternalRef value={issue.externalRef} />}
							</>
						)
					}
				/>

				{!row.isContext && issue.body && (
					<MarkdownPreview className={cn('mt-1 line-clamp-1 text-xs', ISSUE_LINE_INSET)}>{issue.body}</MarkdownPreview>
				)}
			</div>
		</article>
	)
}

type TreeProps = {
	readonly issues: readonly Issue[]
	readonly context: readonly Issue[]
	readonly project: string
	readonly kind: IssueKind
}

const Tree = ({ issues, context, project, kind }: TreeProps) => {
	const rows = buildIssueTree(issues, { context })

	return (
		<div className='border-border border'>
			{rows.map((row, index) => (
				<StaggerItem key={row.issue.id} index={index}>
					<div className={cn(index > 0 && row.depth === 0 && 'border-border border-t')}>
						<Row row={row} project={project} kind={kind} />
					</div>
				</StaggerItem>
			))}
		</div>
	)
}

const GroupedTodos = ({ issues, context, project }: Omit<TreeProps, 'kind'>) => {
	const plans = usePlans(project)
	const tickets = useIssues(project, { kind: ISSUE_KIND.TICKET })

	const groups = groupTodos(
		issues,
		plans.status === Status.Ready ? plans.plans : [],
		tickets.status === Status.Ready ? tickets.issues : [],
	)

	return (
		<div className='space-y-6'>
			{groups.map(({ target, items }) => (
				<section key={target ? `${target.kind}:${target.id}` : 'none'} className='space-y-2'>
					{(target !== undefined || groups.length > 1) && <GroupHeader target={target} slug={project} />}
					<Tree issues={items} context={context} project={project} kind={ISSUE_KIND.TODO} />
				</section>
			))}
		</div>
	)
}

type IssuesProps = {
	readonly kind: IssueKind
	readonly emptyLabel: string
	readonly defaultStatuses?: readonly IssueStatus[]
}

const Issues = ({ kind, emptyLabel, defaultStatuses }: IssuesProps) => {
	const { slug } = useParams({ strict: false }) as { readonly slug: string }
	const search = useSearch({ strict: false }) as IssuesSearch
	const statuses = search.statuses ?? defaultStatuses

	const state = useIssues(slug, {
		kind,
		status: statuses,
		priority: search.priority,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const filtered =
		search.q !== undefined ||
		search.statuses !== undefined ||
		search.tags !== undefined ||
		search.priority !== undefined
	const narrowed = filtered || statuses !== undefined

	const everything = useIssues(slug, { kind, sort: search.sort })

	const list = (() => {
		if (state.status === Status.Loading) return null

		if (state.status === Status.Failed) {
			return <LoadError message={state.message} />
		}

		if (state.issues.length === 0) {
			if (filtered) return <p className='text-dim text-sm'>No {emptyLabel} match.</p>

			return <EmptyState message={`No ${emptyLabel} yet.`} command={`folio ${kind} create "<title>" -p ${slug}`} />
		}

		const context = narrowed && everything.status === Status.Ready ? everything.issues : []

		if (kind === ISSUE_KIND.TODO) {
			return <GroupedTodos issues={state.issues} context={context} project={slug} />
		}

		return <Tree issues={state.issues} context={context} project={slug} kind={kind} />
	})()

	return (
		<div className='space-y-4'>
			<IssueFilters kind={kind} project={slug} defaultStatuses={defaultStatuses} />
			{list}
		</div>
	)
}

export { Issues };

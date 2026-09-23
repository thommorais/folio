import { Link, useParams, useSearch } from '@tanstack/react-router';
import { cn } from '@thom/libs/cn';
import { Badge } from '@thom/ui/badge';
import { useIssues } from '_/app/use-issues';
import { Skeleton } from '_/components/motion/skeleton';
import { StaggerItem } from '_/components/motion/stagger';
import { isTerminal, ISSUE_KIND, ISSUE_STATUS, type IssueKind, type IssueStatus, type Priority } from '_/core/domain/issue';
import { buildIssueTree, type IssueRow } from '_/core/domain/issue-tree';
import type { IssueSortField, Sort } from '_/core/ports/sort';
import { Status } from '_/lib/async-status';
import { useScope } from '_/routing/use-scope';
import { IssueFilters } from './issue-filters';
import { ISSUE_STATUS_LABELS as statusLabels } from './status-labels';

type IssuesSearch = {
	readonly statuses?: readonly IssueStatus[]
	readonly priority?: Priority
	readonly tags?: readonly string[]
	readonly q?: string
	readonly sort?: Sort<IssueSortField>
}

const INDENT = 22
// Vertical center of a nested row's title line, where the connector meets it.
const ELBOW = '1.1rem'

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

	return (
		<article className={cn('relative px-4', row.depth === 0 ? 'py-4' : 'py-3')}>
			<Guides row={row} />

			<div className='space-y-2' style={{ paddingLeft: row.depth * INDENT }}>
				<div className='flex items-start justify-between gap-4'>
					<span className='flex min-w-0 items-start gap-2'>
						{kind === ISSUE_KIND.TODO && (
							<span
								className={cn(
									'border-border mt-0.5 size-4 shrink-0 border',
									issue.status === ISSUE_STATUS.DONE && 'bg-foreground border-foreground',
								)}
							/>
						)}

						<Link
							to='/$client/$domain/$slug/tickets/$ticket'
							params={{ client, domain, slug: project, ticket: issue.slug }}
							className={cn(
								'text-sm font-medium hover:underline',
								(isTerminal(issue.status) || row.isContext) && 'text-dim',
								issue.status === ISSUE_STATUS.DONE && kind === ISSUE_KIND.TODO && 'line-through',
							)}
						>
							{issue.title}
						</Link>
					</span>

					<span className='flex shrink-0 items-center gap-2'>
						{issue.wayfinder && <Badge color='muted'>{issue.wayfinder}</Badge>}
						<span className='text-dim text-xs'>{statusLabels[issue.status]}</span>
					</span>
				</div>

				{/* A context row is only present to place its children, so its own
				    body and metadata would read as a false match. */}
				{!row.isContext && (
					// The checkbox indents the title, so its row's body and metadata
					// line up under the text rather than under the box.
					<div className={cn('space-y-2', kind === ISSUE_KIND.TODO && 'pl-6')}>
						{issue.body && <p className='text-dim line-clamp-2 text-sm'>{issue.body}</p>}

						<div className='flex flex-wrap items-center gap-2 pt-1'>
							<span className='text-dimmer font-mono text-xs'>{issue.priority}</span>
							{issue.externalRef && <span className='text-dimmer font-mono text-xs'>{issue.externalRef}</span>}
							{issue.tags.map(tag => (
								<Badge key={tag} color='muted'>
									{tag}
								</Badge>
							))}
						</div>
					</div>
				)}
			</div>
		</article>
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
		statuses !== undefined ||
		search.tags !== undefined ||
		search.priority !== undefined

	const everything = useIssues(slug, { kind, sort: search.sort })

	const list = (() => {
		if (state.status === Status.Loading) {
			return (
				<div className='border-border divide-border divide-y border'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='h-20' />
					))}
				</div>
			)
		}

		if (state.status === Status.Failed) {
			return <p className='text-destructive text-sm'>{state.message}</p>
		}

		if (state.issues.length === 0) {
			return <p className='text-dim text-sm'>{filtered ? `No ${emptyLabel} match.` : `No ${emptyLabel} yet.`}</p>
		}

		const context = filtered && everything.status === Status.Ready ? everything.issues : []
		const rows = buildIssueTree(state.issues, { context })

		return (
			<div className='border-border border'>
				{rows.map((row, index) => (
					<StaggerItem key={row.issue.id} index={index}>
						<div className={cn(index > 0 && row.depth === 0 && 'border-border border-t')}>
							<Row row={row} project={slug} kind={kind} />
						</div>
					</StaggerItem>
				))}
			</div>
		)
	})()

	return (
		<div className='space-y-4'>
			<IssueFilters defaultStatuses={defaultStatuses} />
			{list}
		</div>
	)
}

export { Issues };

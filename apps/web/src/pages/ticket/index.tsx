import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { Badge } from '@thom/ui/badge';
import { Tag } from '_/components/issue/tag';
import { IssueLine } from '_/components/issue/issue-line';
import { Flash } from '_/components/motion/flash';
import { CycleChip } from '_/components/issue/cycle-chip';
import { PriorityIcon } from '_/components/issue/priority-icon';
import { StatusIcon } from '_/components/issue/status-icon';
import { Heading } from '@thom/ui/heading';
import { useCycles } from '_/app/use-cycles';
import { useEntries } from '_/app/use-entries';
import { useIssue } from '_/app/use-issue';
import { useIssues } from '_/app/use-issues';
import { usePlans } from '_/app/use-plans';
import { Markdown } from '_/components/markdown'
import { WorkLog } from '_/components/record/work-log';
import { RecordGone } from '_/components/record/record-gone';
import { ShareSheet } from '_/components/share/share-sheet';
import { openBlockers } from '_/core/domain/blocked';
import { newestFirst } from '_/core/domain/cycle-progress';
import { ADDRESSABLE_KINDS } from '_/core/domain/entry';
import { ISSUE_KIND, type Issue, type IssueStatus } from '_/core/domain/issue';
import { Status } from '_/lib/async-status';
import { ENTRY_KIND_LABELS } from '_/pages/journal/kind-labels';
import { useScope } from '_/routing/use-scope';
import { useSlugSync } from '_/routing/use-slug-sync';
import { Answer } from './answer';
import { CycleTimeline } from './cycle-timeline';
import { MapFrontier } from './map-frontier';
import { SHARE_KIND } from '_/core/domain/share';
import { LoadError } from '_/components/load-error';
import { ExternalRef } from '_/components/issue/external-ref';
import { CopyId } from '_/components/issue/copy-id';

const statusLabels: Record<IssueStatus, string> = {
	open: 'Open',
	in_progress: 'In progress',
	blocked: 'Blocked',
	done: 'Done',
	cancelled: 'Cancelled',
}

// Each section lists the ticket's own slice of an entity. The hooks take the
// ticket id, so a ticket with no work of a given kind renders nothing rather
// than the project's whole list.
const Section = ({ title, children }: { readonly title: string; readonly children: React.ReactNode }) => (
	<section className='space-y-2'>
		<h2 className='text-base font-medium'>{title}</h2>
		{children}
	</section>
)

const Empty = ({ what }: { readonly what: string }) => <p className='text-dim text-sm'>No {what} on this ticket.</p>

const TicketDetail = () => {
	const {
		client,
		domain,
		slug,
		ticket: ticketSlug,
	} = useParams({ from: '/_authenticated/$client/$domain/$slug/tickets/$ticket' })
	const state = useIssue(slug, ticketSlug)

	const navigate = useNavigate()

	useSlugSync({
		current: ticketSlug,
		record: state.status === Status.Ready ? state.issue : undefined,
		rename: renamed =>
			void navigate({
				to: '/$client/$domain/$slug/tickets/$ticket',
				params: { client, domain, slug, ticket: renamed },
				replace: true,
			}),
	})

	if (state.status === Status.Idle || state.status === Status.Loading) {
		return null
	}

	if (state.status === Status.Gone) {
		return (
			<RecordGone title={state.title}>
				<Link to='/$client/$domain/$slug/tickets' params={{ client, domain, slug }} className='text-sm underline'>
					Back to tickets
				</Link>
			</RecordGone>
		)
	}

	if (state.status === Status.Failed) {
		return <LoadError message={state.message} />
	}

	return <TicketBody project={slug} ticket={state.issue} />
}

type BodyProps = {
	readonly project: string
	readonly ticket: Issue
}

const TicketBody = ({ project, ticket }: BodyProps) => {
	const { client, domain } = useScope()
	const ticketId = ticket.id
	const plans = usePlans(project, { ticketId })
	const todos = useIssues(project, { kind: ISSUE_KIND.TODO, parentId: ticketId })
	const journal = useEntries(project, { kinds: ADDRESSABLE_KINDS, issueId: ticketId })
	const cycles = useCycles(project, { ticketId })
	const planning = useCycles(project, { mapId: ticketId })
	const planned = planning.status === Status.Ready ? newestFirst(planning.cycles).at(0) : undefined
	const childTickets = useIssues(project, { kind: ISSUE_KIND.TICKET, parentId: ticketId })
	const maps =
		childTickets.status === Status.Ready ? childTickets.issues.filter(child => child.wayfinder === 'map') : []
	const current = cycles.status === Status.Ready ? newestFirst(cycles.cycles).at(0) : undefined
	const siblings = useIssues(project)
	const parent =
		ticket.parentId !== undefined && siblings.status === Status.Ready
			? siblings.issues.find(candidate => candidate.id === ticket.parentId)
			: undefined
	const waiting = siblings.status === Status.Ready ? openBlockers(ticket, siblings.issues) : 0

	return (
		<div className='space-y-8'>
			<header className='space-y-3'>
				<div className='flex items-start justify-between gap-4'>
					<Heading>{ticket.title}</Heading>
					<div className='flex shrink-0 items-center gap-2'>
						<CopyId id={ticket.id} />
						<ShareSheet target={{ kind: SHARE_KIND.ISSUE, id: ticket.id, projectId: ticket.projectId }} />
					</div>
				</div>

				<div className='flex flex-wrap items-center gap-2'>
					<span className='text-dim flex items-center gap-1.5 text-xs'>
						<StatusIcon status={ticket.status} />
						{statusLabels[ticket.status]}
					</span>
					<span className='text-dim flex items-center gap-1.5 text-xs capitalize'>
						<PriorityIcon priority={ticket.priority} />
						{ticket.priority}
					</span>
					{ticket.wayfinder && <Badge color='neutral'>{ticket.wayfinder}</Badge>}
					{current && <CycleChip cycle={current} />}
					{ticket.externalRef && <ExternalRef value={ticket.externalRef} />}
					{ticket.tags.map(tag => (
						<Tag key={tag} tag={tag} />
					))}
				</div>

				{ticket.body && <Markdown>{ticket.body}</Markdown>}

				{ticket.resolution && <Answer project={project} ticket={ticket} />}

				{(parent !== undefined || waiting > 0) && (
					<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
						{parent !== undefined && (
							<span>
								{planned !== undefined && planned.ticketId === parent.id ? `plans cycle ${planned.ordinal} of ` : 'under '}
								<Link
									to='/$client/$domain/$slug/tickets/$ticket'
									params={{ client, domain, slug: project, ticket: parent.slug }}
									className='hover:text-foreground underline underline-offset-2 transition-colors'
								>
									{parent.title}
								</Link>
							</span>
						)}
						{waiting > 0 && <span>waits on {waiting}</span>}
					</div>
				)}
			</header>

			{ticket.wayfinder === 'map' && <MapFrontier project={project} map={ticket} />}

			{cycles.status === Status.Ready && cycles.cycles.length > 0 && (
				<Section title='Cycles'>
					<CycleTimeline project={project} cycles={cycles.cycles} maps={maps} />
				</Section>
			)}

			<Section title='Plans'>
				{plans.status === Status.Ready && plans.plans.length === 0 && <Empty what='plans' />}
				{plans.status === Status.Ready && plans.plans.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{plans.plans.map(plan => (
							<li key={plan.id} className='flex items-center justify-between gap-4 px-4 py-3 text-sm'>
								<span>{plan.title}</span>
								<span className='text-dim shrink-0 text-xs'>{plan.status}</span>
							</li>
						))}
					</ul>
				)}
			</Section>

			<Section title='Todos'>
				{todos.status === Status.Ready && todos.issues.length === 0 && <Empty what='todos' />}
				{todos.status === Status.Ready && todos.issues.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{todos.issues.map(todo => (
							<li key={todo.id}>
								<Link
									to='/$client/$domain/$slug/todos/$todo'
									params={{ client, domain, slug: project, todo: todo.slug }}
									data-row
									className='hover:bg-accent/40 active:bg-accent/60 relative block px-4 py-2.5 transition-colors'
								>
									<Flash on={todo.updatedAt.getTime()} />
									<IssueLine issue={todo} strike />
								</Link>
							</li>
						))}
					</ul>
				)}
			</Section>

			<Section title='Journal'>
				{journal.status === Status.Ready && journal.entries.length === 0 && <Empty what='entries' />}
				{journal.status === Status.Ready && journal.entries.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{journal.entries.map(entry => (
							<li key={entry.id}>
								<Link
									to='/$client/$domain/$slug/journal/$entry'
									params={{ client, domain, slug: project, entry: entry.slug }}
									data-row
									className='hover:bg-accent/40 active:bg-accent/60 flex items-center justify-between gap-4 px-4 py-3 text-sm transition-colors'
								>
									<span className='min-w-0 flex-1 truncate'>{entry.title}</span>
									<Badge color='muted'>{ENTRY_KIND_LABELS[entry.kind]}</Badge>
								</Link>
							</li>
						))}
					</ul>
				)}
			</Section>

			<WorkLog project={project} issueId={ticketId} />
		</div>
	)
}

export { TicketBody, TicketDetail };

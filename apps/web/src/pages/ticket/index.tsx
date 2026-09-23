import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { cn } from '@thom/libs/cn';
import { Badge } from '@thom/ui/badge';
import { Heading } from '@thom/ui/heading';
import { useCycles } from '_/app/use-cycles';
import { useEntries } from '_/app/use-entries';
import { useIssue } from '_/app/use-issue';
import { useIssues } from '_/app/use-issues';
import { usePlans } from '_/app/use-plans';
import { Markdown } from '_/components/markdown';
import { RecordGone } from '_/components/record/record-gone';
import { ShareSheet } from '_/components/share/share-sheet';
import { newestFirst } from '_/core/domain/cycle-progress';
import { ADDRESSABLE_KINDS, KIND } from '_/core/domain/entry';
import type { Issue, IssueStatus } from '_/core/domain/issue';
import { Status } from '_/lib/async-status';
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels';
import { ENTRY_KIND_LABELS } from '_/pages/journal/kind-labels';
import { useScope } from '_/routing/use-scope';
import { useSlugSync } from '_/routing/use-slug-sync';
import { CycleTimeline } from './cycle-timeline';
import { MapFrontier } from './map-frontier';

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
		<h2 className='text-dim text-xs tracking-wide uppercase'>{title}</h2>
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
		return <div className='bg-accent/40 h-32 animate-pulse' />
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
		return <p className='text-destructive text-sm'>{state.message}</p>
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
	const todos = useIssues(project, { kind: KIND.TODO, parentId: ticketId })
	const journal = useEntries(project, { kinds: ADDRESSABLE_KINDS, issueId: ticketId })
	const cycles = useCycles(project, { ticketId })
	const workLog = useEntries(project, { kind: KIND.LOG, issueId: ticketId })

	// A log stamped with a cycle is shown on that round in the timeline, so
	// only the loose ones are left for the section below it.
	const unstamped = workLog.status === Status.Ready ? workLog.entries.filter(entry => entry.cycleId === undefined) : []

	const current = cycles.status === Status.Ready ? newestFirst(cycles.cycles).at(0) : undefined
	const siblings = useIssues(project, { kind: KIND.TICKET})
	const parent =
		ticket.parentId !== undefined && siblings.status === Status.Ready
			? siblings.issues.find(candidate => candidate.id === ticket.parentId)
			: undefined

	return (
		<div className='space-y-8'>
			<header className='space-y-3'>
				<div className='flex items-start justify-between gap-4'>
					<Heading>{ticket.title}</Heading>
					<ShareSheet target={{ kind: 'issue', id: ticket.id, projectId: ticket.projectId }} />
				</div>

				<div className='flex flex-wrap items-center gap-2'>
					<span className='text-dim text-xs'>{statusLabels[ticket.status]}</span>
					<span className='text-dimmer font-mono text-xs'>{ticket.priority}</span>
					{ticket.wayfinder && <Badge color='neutral'>{ticket.wayfinder}</Badge>}
					{current && (
						<Badge color='muted'>
							cycle {current.ordinal}: {current.phase}
						</Badge>
					)}
					{ticket.externalRef && <span className='text-dimmer font-mono text-xs'>{ticket.externalRef}</span>}
					{ticket.tags.map(tag => (
						<Badge key={tag} color='muted'>
							{tag}
						</Badge>
					))}
				</div>

				{ticket.body && <Markdown>{ticket.body}</Markdown>}

				{(parent !== undefined || ticket.dependsOn.length > 0) && (
					<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
						{parent !== undefined && (
							<span>
								under{' '}
								<Link
									to='/$client/$domain/$slug/tickets/$ticket'
									params={{ client, domain, slug: project, ticket: parent.slug }}
									className='hover:text-foreground underline underline-offset-2 transition-colors'
								>
									{parent.title}
								</Link>
							</span>
						)}
						{ticket.dependsOn.length > 0 && <span>waits on {ticket.dependsOn.length}</span>}
					</div>
				)}
			</header>

			{ticket.wayfinder === 'map' && <MapFrontier project={project} map={ticket} />}

			{cycles.status === Status.Ready && cycles.cycles.length > 0 && (
				<Section title='Cycles'>
					<CycleTimeline project={project} cycles={cycles.cycles} />
				</Section>
			)}

			{/* Logs written outside a cycle have no round to sit under, so they
			    keep a section of their own. */}
			{unstamped.length > 0 && (
				<Section title='Work log'>
					<ul className='border-border divide-border divide-y border'>
						{unstamped.map(entry => (
							<li key={entry.id} className='space-y-1 px-4 py-3'>
								<p className='text-sm whitespace-pre-line'>{entry.body}</p>
								<p className='text-dimmer text-xs'>{entry.createdAt.toLocaleString()}</p>
							</li>
						))}
					</ul>
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
									to='/$client/$domain/$slug/todos'
									params={{ client, domain, slug: project }}
									search={{ todo: todo.id }}
									className='hover:bg-accent/40 flex w-full items-center gap-3 px-4 py-3 transition-colors'
								>
									<span
										className={cn(
											'border-border size-4 shrink-0 border',
											todo.status === 'done' && 'bg-foreground border-foreground',
										)}
									/>

									<span className={cn('flex-1 truncate text-sm', todo.status === 'done' && 'text-dim line-through')}>
										{todo.title}
									</span>

									{todo.status === 'blocked' && <Badge color='destructive'>Blocked</Badge>}

									<span className='hidden shrink-0 items-center gap-3 sm:flex'>
										{todo.tags.map(tag => (
											<Badge key={tag} color='muted'>
												{tag}
											</Badge>
										))}
									</span>

									<span className='text-dimmer hidden w-16 shrink-0 text-right text-xs sm:block'>{todo.priority}</span>
									<span className='text-dim w-24 shrink-0 text-right text-xs'>{ISSUE_STATUS_LABELS[todo.status]}</span>
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
									className='hover:bg-accent/40 flex items-center justify-between gap-4 px-4 py-3 text-sm transition-colors'
								>
									<span className='min-w-0 flex-1 truncate'>{entry.title}</span>
									<Badge color='muted'>{ENTRY_KIND_LABELS[entry.kind]}</Badge>
								</Link>
							</li>
						))}
					</ul>
				)}
			</Section>
		</div>
	)
}

export { TicketDetail };

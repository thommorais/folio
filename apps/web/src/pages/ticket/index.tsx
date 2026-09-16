import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { cn } from '@thom/libs/cn';
import { Badge } from '@thom/ui/badge';
import { Heading } from '@thom/ui/heading';
import { useCycles } from '_/app/use-cycles';
import { useDocs } from '_/app/use-docs';
import { useJournal } from '_/app/use-journal';
import { usePlans } from '_/app/use-plans';
import { useTicket } from '_/app/use-ticket';
import { useTicketLogs } from '_/app/use-ticket-logs';
import { useTickets } from '_/app/use-tickets';
import { useTodos } from '_/app/use-todos';
import { Markdown } from '_/components/markdown';
import { RecordGone } from '_/components/record/record-gone';
import { isResolved } from '_/core/domain/cycle';
import type { TicketStatus } from '_/core/domain/ticket';
import { TODO_STATUS_LABELS } from '_/pages/todos/status-labels';
import { useSlugSync } from '_/routing/use-slug-sync';
import { MapFrontier } from './map-frontier';

const statusLabels: Record<TicketStatus, string> = {
	open: 'Open',
	in_progress: 'In progress',
	blocked: 'Blocked',
	closed: 'Closed',
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
	const { slug, ticket: ticketSlug } = useParams({ from: '/_authenticated/$slug/tickets/$ticket' })
	const state = useTicket(slug, ticketSlug)

	const navigate = useNavigate()

	useSlugSync({
		current: ticketSlug,
		record: state.status === 'ready' ? state.ticket : undefined,
		rename: renamed =>
			void navigate({ to: '/$slug/tickets/$ticket', params: { slug, ticket: renamed }, replace: true }),
	})

	if (state.status === 'idle' || state.status === 'loading') {
		return <div className='bg-accent/40 h-32 animate-pulse' />
	}

	if (state.status === 'gone') {
		return (
			<RecordGone title={state.title}>
				<Link to='/$slug/tickets' params={{ slug }} className='text-sm underline'>
					Back to tickets
				</Link>
			</RecordGone>
		)
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return <TicketBody project={slug} ticket={state.ticket} />
}

type BodyProps = {
	readonly project: string
	readonly ticket: import('_/core/domain/ticket').Ticket
}

const TicketBody = ({ project, ticket }: BodyProps) => {
	const ticketId = ticket.id
	const plans = usePlans(project, { ticketId })
	const todos = useTodos(project, { ticketId })
	const journal = useJournal(project, { ticketId })
	const docs = useDocs(project, { ticketId })
	const cycles = useCycles(project, { ticketId })
	const workLog = useTicketLogs(project, { ticketId })

	const current = cycles.status === 'ready' ? cycles.cycles.at(-1) : undefined
	const siblings = useTickets(project, {})
	const parent =
		ticket.parentId !== undefined && siblings.status === 'ready'
			? siblings.tickets.find(candidate => candidate.id === ticket.parentId)
			: undefined

	return (
		<div className='space-y-8'>
			<header className='space-y-3'>
				<Heading>{ticket.title}</Heading>

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

				{ticket.body &&<Markdown>{ticket.body}</Markdown>}

				{(parent !== undefined || ticket.dependsOn.length > 0) && (
					<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
						{parent !== undefined && (
							<span>
								under{' '}
								<Link
									to='/$slug/tickets/$ticket'
									params={{ slug: project, ticket: parent.slug }}
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

			{ticket.wayfinder === 'map' && <MapFrontier project={project} mapId={ticketId} />}

			{cycles.status === 'ready' && cycles.cycles.length > 0 && (
				<Section title='Cycles'>
					<ul className='border-border divide-border divide-y border'>
						{cycles.cycles.map(cycle => (
							<li key={cycle.id} className='space-y-1 px-4 py-3 text-sm'>
								<div className='flex items-center gap-2'>
									<span className='text-dimmer font-mono text-xs'>{cycle.ordinal}</span>
									<span>{cycle.phase}</span>
									{isResolved(cycle) ? <Badge color='muted'>resolved</Badge> : <Badge color='active'>open</Badge>}
								</div>
								{cycle.resolution && <p className='text-dim text-xs'>{cycle.resolution}</p>}
							</li>
						))}
					</ul>
				</Section>
			)}

			{workLog.status === 'ready' && workLog.logs.length > 0 && (
				<Section title='Work log'>
					<ul className='border-border divide-border divide-y border'>
						{workLog.logs.map(entry => (
							<li key={entry.id} className='space-y-1 px-4 py-3'>
								<p className='text-sm whitespace-pre-line'>{entry.body}</p>
								<p className='text-dimmer text-xs'>{entry.createdAt.toLocaleString()}</p>
							</li>
						))}
					</ul>
				</Section>
			)}

			<Section title='Plans'>
				{plans.status === 'ready' && plans.plans.length === 0 && <Empty what='plans' />}
				{plans.status === 'ready' && plans.plans.length > 0 && (
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
				{todos.status === 'ready' && todos.todos.length === 0 && <Empty what='todos' />}
				{todos.status === 'ready' && todos.todos.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{todos.todos.map(todo => (
							<li key={todo.id}>
								<Link
									to='/$slug/todos'
									params={{ slug: project }}
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
									<span className='text-dim w-24 shrink-0 text-right text-xs'>{TODO_STATUS_LABELS[todo.status]}</span>
								</Link>
							</li>
						))}
					</ul>
				)}
			</Section>

			<Section title='Journal'>
				{journal.status === 'ready' && journal.journal.length === 0 && <Empty what='journal entries' />}
				{journal.status === 'ready' && journal.journal.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{journal.journal.map(entry => (
							<li key={entry.id} className='px-4 py-3 text-sm'>
								{entry.title}
							</li>
						))}
					</ul>
				)}
			</Section>

			<Section title='Docs'>
				{docs.status === 'ready' && docs.docs.length === 0 && <Empty what='docs' />}
				{docs.status === 'ready' && docs.docs.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{docs.docs.map(doc => (
							<li key={doc.id} className='flex items-center justify-between gap-4 px-4 py-3 text-sm'>
								<span>{doc.title}</span>
								<span className='text-dimmer font-mono text-xs'>{doc.slug}</span>
							</li>
						))}
					</ul>
				)}
			</Section>
		</div>
	)
}

export { TicketDetail };

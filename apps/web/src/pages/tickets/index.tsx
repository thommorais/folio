import { Link, useParams, useSearch } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useTickets } from '_/app/use-tickets'
import { isTerminal } from '_/core/domain/ticket'
import { buildTicketTree, type TicketRow } from '_/core/domain/ticket-tree'
import { TicketFilters } from './ticket-filters'
import { TICKET_STATUS_LABELS as statusLabels } from './status-labels'

const INDENT = 22
// Vertical center of a nested row's title line, where the connector meets it.
const ELBOW = '1.1rem'

// Guides live in the row's left gutter rather than in the flow, so the title
// column keeps one offset per depth instead of drifting with the markup.
const Guides = ({ row }: { readonly row: TicketRow }) => {
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

const Row = ({ row, project }: { readonly row: TicketRow; readonly project: string }) => {
	const { ticket } = row

	return (
		<article className={cn('relative px-4', row.depth === 0 ? 'py-4' : 'py-3')}>
			<Guides row={row} />

			<div className='space-y-2' style={{ paddingLeft: row.depth * INDENT }}>
				<div className='flex items-start justify-between gap-4'>
					<Link
						to='/$slug/tickets/$ticket'
						params={{ slug: project, ticket: ticket.slug }}
						className={cn(
							'text-sm font-medium hover:underline',
							(isTerminal(ticket.status) || row.isContext) && 'text-dim',
						)}
					>
						{ticket.title}
					</Link>

					<span className='flex shrink-0 items-center gap-2'>
						{ticket.wayfinder && <Badge color='muted'>{ticket.wayfinder}</Badge>}
						<span className='text-dim text-xs'>{statusLabels[ticket.status]}</span>
					</span>
				</div>

				{/* A context row is only present to place its children, so its own
				    body and metadata would read as a false match. */}
				{!row.isContext && (
					<>
						{ticket.body && <p className='text-dim line-clamp-2 text-sm'>{ticket.body}</p>}

						<div className='flex flex-wrap items-center gap-2 pt-1'>
							<span className='text-dimmer font-mono text-xs'>{ticket.priority}</span>
							{ticket.externalRef && <span className='text-dimmer font-mono text-xs'>{ticket.externalRef}</span>}
							{ticket.tags.map(tag => (
								<Badge key={tag} color='muted'>
									{tag}
								</Badge>
							))}
						</div>
					</>
				)}
			</div>
		</article>
	)
}

const Tickets = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/tickets/' })
	const search = useSearch({ from: '/_authenticated/$slug/tickets/' })

	const state = useTickets(slug, {
		status: search.statuses,
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

	// A filtered list carries matches only, so the ancestors needed to place them
	// come from an unfiltered read. The adapter fetches the project's tickets
	// whole either way, so this is the same query the list already makes.
	const everything = useTickets(slug, { sort: search.sort })

	const list = (() => {
		if (state.status === 'loading') {
			return (
				<div className='border-border divide-border divide-y border'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='h-20' />
					))}
				</div>
			)
		}

		if (state.status === 'failed') {
			return <p className='text-destructive text-sm'>{state.message}</p>
		}

		if (state.tickets.length === 0) {
			return <p className='text-dim text-sm'>{filtered ? 'No tickets match.' : 'No tickets yet.'}</p>
		}

		const context = filtered && everything.status === 'ready' ? everything.tickets : []
		const rows = buildTicketTree(state.tickets, { context })

		return (
			<div className='border-border border'>
				{rows.map((row, index) => (
					<StaggerItem key={row.ticket.id} index={index}>
						{/* Only roots get a rule. Nested rows are already separated by
						    their connector, and a full-width border would cut across it. */}
						<div className={cn(index > 0 && row.depth === 0 && 'border-border border-t')}>
							<Row row={row} project={slug} />
						</div>
					</StaggerItem>
				))}
			</div>
		)
	})()

	return (
		<div className='space-y-4'>
			<TicketFilters />
			{list}
		</div>
	)
}

export { Tickets }

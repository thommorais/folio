import { Link, useParams, useSearch } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useTickets } from '_/app/use-tickets'
import { isTerminal } from '_/core/domain/ticket'
import { TicketFilters } from './ticket-filters'
import { TICKET_STATUS_LABELS as statusLabels } from './status-labels'

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

		return (
			<div className='border-border divide-border divide-y border'>
				{state.tickets.map((ticket, index) => (
					<StaggerItem key={ticket.id} index={index}>
						<article className='space-y-2 px-4 py-4'>
							<div className='flex items-start justify-between gap-4'>
								<Link
									to='/$slug/tickets/$ticket'
									params={{ slug, ticket: ticket.slug }}
									className={cn('text-sm font-medium hover:underline', isTerminal(ticket.status) && 'text-dim')}
								>
									{ticket.title}
								</Link>
								<span className='text-dim shrink-0 text-xs'>{statusLabels[ticket.status]}</span>
							</div>

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
						</article>
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

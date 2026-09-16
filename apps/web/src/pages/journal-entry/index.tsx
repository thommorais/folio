import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { useSlugSync } from '_/routing/use-slug-sync'
import { RecordGone } from '_/components/record/record-gone'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Markdown } from '_/components/markdown'
import { useJournalEntry } from '_/app/use-journal-entry'
import { useTickets } from '_/app/use-tickets'
import type { JournalEntry } from '_/core/domain/journal'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const EntryBody = ({ project, entry }: { readonly project: string; readonly entry: JournalEntry }) => {
	const tickets = useTickets(project, {})
	const ticket =
		entry.ticketId !== undefined && tickets.status === 'ready'
			? tickets.tickets.find(candidate => candidate.id === entry.ticketId)
			: undefined

	return (
		<article className='space-y-8'>
			<header className='space-y-3'>
				<Heading>{entry.title}</Heading>

				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<span>{formatDate(entry.createdAt)}</span>
					{entry.branch && <span className='font-mono'>{entry.branch}</span>}
					{entry.pr && <span className='font-mono'>#{entry.pr}</span>}
					{entry.externalRef && <span className='font-mono'>{entry.externalRef}</span>}
					{entry.tags.map(tag => (
						<Badge key={tag} color='muted'>
							{tag}
						</Badge>
					))}
				</div>

				{ticket !== undefined && (
					<div className='text-dimmer text-xs'>
						<span>
							under{' '}
							<Link
								to='/$slug/tickets/$ticket'
								params={{ slug: project, ticket: ticket.slug }}
								className='hover:text-foreground underline underline-offset-2 transition-colors'
							>
								{ticket.title}
							</Link>
						</span>
					</div>
				)}
			</header>

			{entry.body ? <Markdown>{entry.body}</Markdown> : <p className='text-dim text-sm'>This entry has no body.</p>}
		</article>
	)
}

const JournalEntryDetail = () => {
	const { slug, entry } = useParams({ from: '/_authenticated/$slug/journal/$entry' })
	const state = useJournalEntry(slug, entry)

	const navigate = useNavigate()

	useSlugSync({
		current: entry,
		record: state.status === 'ready' ? state.entry : undefined,
		rename: renamed => void navigate({ to: '/$slug/journal/$entry', params: { slug, entry: renamed }, replace: true }),
	})

	if (state.status === 'idle' || state.status === 'loading') {
		return <div className='bg-accent/40 h-32 animate-pulse' />
	}

	if (state.status === 'gone') {
		return (
			<RecordGone title={state.title}>
				<Link to='/$slug/journal' params={{ slug }} className='text-sm underline'>
					Back to the journal
				</Link>
			</RecordGone>
		)
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return <EntryBody project={slug} entry={state.entry} />
}

export { JournalEntryDetail }

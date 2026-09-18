import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { useSlugSync } from '_/routing/use-slug-sync'
import { RecordGone } from '_/components/record/record-gone'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Markdown } from '_/components/markdown'
import { useEntry } from '_/app/use-entry'
import { useIssues } from '_/app/use-issues'
import type { Entry } from '_/core/domain/entry'
import { useScope } from '_/routing/use-scope'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const EntryBody = ({ project, entry }: { readonly project: string; readonly entry: Entry }) => {
	const { client, domain } = useScope()
	const tickets = useIssues(project, {})
	const ticket =
		entry.issueId !== undefined && tickets.status === 'ready'
			? tickets.issues.find(candidate => candidate.id === entry.issueId)
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
								to='/$client/$domain/$slug/tickets/$ticket'
								params={{ client, domain, slug: project, ticket: ticket.slug }}
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
	const { client, domain, slug, entry } = useParams({ from: '/_authenticated/$client/$domain/$slug/journal/$entry' })
	const state = useEntry(slug, entry)

	const navigate = useNavigate()

	useSlugSync({
		current: entry,
		record: state.status === 'ready' ? state.entry : undefined,
		rename: renamed =>
			void navigate({
				to: '/$client/$domain/$slug/journal/$entry',
				params: { client, domain, slug, entry: renamed },
				replace: true,
			}),
	})

	if (state.status === 'idle' || state.status === 'loading') {
		return <div className='bg-accent/40 h-32 animate-pulse' />
	}

	if (state.status === 'gone') {
		return (
			<RecordGone title={state.title}>
				<Link to='/$client/$domain/$slug/journal' params={{ client, domain, slug }} className='text-sm underline'>
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

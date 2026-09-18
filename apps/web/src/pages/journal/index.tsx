import { Link, useParams, useSearch } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useEntries } from '_/app/use-entries'
import { ADDRESSABLE_KINDS } from '_/core/domain/entry'
import { JournalFilters } from './journal-filters'
import { ENTRY_KIND_LABELS } from './kind-labels'

const dayMonth = new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short' })

const Journal = () => {
	const { client, domain, slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/journal/' })
	const state = useEntries(slug, {
		// Without a selection the list is both kinds, never the ticket work logs.
		kinds: search.kinds ?? ADDRESSABLE_KINDS,
		issueId: search.ticket,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const filtered =
		search.q !== undefined || search.kinds !== undefined || search.tags !== undefined || search.ticket !== undefined

	const list = (() => {
		if (state.status === 'loading') {
			return (
				<div className='border-border divide-border divide-y border'>
					{[0, 1, 2].map(key => (
						<Skeleton key={key} className='h-24' />
					))}
				</div>
			)
		}

		if (state.status === 'failed') {
			return <p className='text-destructive text-sm'>{state.message}</p>
		}

		if (state.entries.length === 0) {
			return <p className='text-dim text-sm'>{filtered ? 'No entries match.' : 'No entries yet.'}</p>
		}

		return (
			<div className='border-border divide-border divide-y border'>
				{state.entries.map((entry, index) => (
					<StaggerItem key={entry.id} index={index}>
						<Link
							to='/$client/$domain/$slug/journal/$entry'
							params={{ client, domain, slug, entry: entry.slug }}
							className='hover:bg-accent/40 block space-y-2 px-4 py-4 transition-colors'
						>
							<div className='flex items-start justify-between gap-4'>
								<h3 className='text-sm font-medium'>{entry.title}</h3>
								<span className='flex shrink-0 items-center gap-3'>
									<Badge color='muted'>{ENTRY_KIND_LABELS[entry.kind]}</Badge>
									<span className='text-dimmer text-xs'>{dayMonth.format(entry.createdAt)}</span>
								</span>
							</div>

							{entry.body && <p className='text-dim line-clamp-2 text-sm'>{entry.body}</p>}

							<div className='flex flex-wrap items-center gap-2 pt-1'>
								{entry.branch && <span className='text-dimmer font-mono text-xs'>{entry.branch}</span>}
								{entry.externalRef && <span className='text-dimmer font-mono text-xs'>{entry.externalRef}</span>}
								{entry.pr && <span className='text-dimmer font-mono text-xs'>#{entry.pr}</span>}
								{entry.tags.map(tag => (
									<Badge key={tag} color='muted'>
										{tag}
									</Badge>
								))}
							</div>
						</Link>
					</StaggerItem>
				))}
			</div>
		)
	})()

	return (
		<div className='space-y-4'>
			<JournalFilters />
			{list}
		</div>
	)
}

export { Journal }

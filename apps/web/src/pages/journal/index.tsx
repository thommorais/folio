import { Link, useParams, useSearch } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Tag } from '_/components/issue/tag'
import { MarkdownPreview } from '_/components/markdown/preview'
import { StaggerItem } from '_/components/motion/stagger'
import { Flash } from '_/components/motion/flash'
import { EmptyState } from '_/components/empty-state'
import { useEntries } from '_/app/use-entries'
import { ADDRESSABLE_KINDS } from '_/core/domain/entry'
import { JournalFilters } from './journal-filters'
import { ENTRY_KIND_LABELS } from './kind-labels'
import { Status } from '_/lib/async-status'
import { LoadError } from '_/components/load-error'

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
		if (state.status === Status.Loading) return null

		if (state.status === Status.Failed) {
			return <LoadError message={state.message} />
		}

		if (state.entries.length === 0) {
			if (filtered) return <p className='text-dim text-sm'>No entries match.</p>

			return <EmptyState message='No entries yet.' command={`folio journal write "<title>" -p ${slug}`} />
		}

		return (
			<div className='border-border divide-border divide-y border'>
				{state.entries.map((entry, index) => (
					<StaggerItem key={entry.id} index={index}>
						<Link
							to='/$client/$domain/$slug/journal/$entry'
							params={{ client, domain, slug, entry: entry.slug }}
							data-row
							className='hover:bg-accent/40 active:bg-accent/60 relative block space-y-2 px-4 py-4 transition-colors'
						>
							<Flash on={entry.updatedAt.getTime()} />
							<div className='flex items-start justify-between gap-4'>
								<h3 className='text-sm font-medium'>{entry.title}</h3>
								<span className='flex shrink-0 items-center gap-3'>
									<Badge color='muted'>{ENTRY_KIND_LABELS[entry.kind]}</Badge>
									<span className='text-dimmer text-xs tabular-nums'>{dayMonth.format(entry.createdAt)}</span>
								</span>
							</div>

							{entry.body && <MarkdownPreview>{entry.body}</MarkdownPreview>}

							<div className='flex flex-wrap items-center gap-2 pt-1'>
								{entry.branch && <span className='text-dimmer font-mono text-xs'>{entry.branch}</span>}
								{entry.externalRef && <span className='text-dimmer font-mono text-xs'>{entry.externalRef}</span>}
								{entry.pr && <span className='text-dimmer font-mono text-xs'>#{entry.pr}</span>}
								{entry.tags.map(tag => (
									<Tag key={tag} tag={tag} />
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

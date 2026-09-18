import { Link, useParams, useSearch } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useEntries } from '_/app/use-entries'
import { DocsFilters } from './doc-filters'

const Docs = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/docs/' })
	const search = useSearch({ from: '/_authenticated/$slug/docs/' })
	const state = useEntries(slug, {
		kind: 'doc',
		issueId: search.ticket,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const filtered = search.q !== undefined || search.tags !== undefined || search.ticket !== undefined

	const list = (() => {
		if (state.status === 'loading') {
			return (
				<div className='grid gap-4 sm:grid-cols-2'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='border-border h-32 border' />
					))}
				</div>
			)
		}

		if (state.status === 'failed') {
			return <p className='text-destructive text-sm'>{state.message}</p>
		}

		if (state.entries.length === 0) {
			return <p className='text-dim text-sm'>{filtered ? 'No docs match.' : 'No docs yet.'}</p>
		}

		return (
			<div className='grid gap-4 sm:grid-cols-2'>
				{state.entries.map((doc, index) => (
					<StaggerItem key={doc.id} index={index}>
						<Link to='/$slug/docs/$doc' params={{ slug, doc: doc.slug }} className='block'>
							<Card interactive>
								<CardHeader>
									<div className='flex items-start justify-between gap-4'>
										<CardTitle>{doc.title}</CardTitle>
										{doc.tags.map(tag => (
											<Badge key={tag} color='muted'>
												{tag}
											</Badge>
										))}
									</div>

									<CardDescription>{doc.body.slice(0, 140) || 'Empty.'}</CardDescription>
									<span className='text-dimmer pt-2 text-xs'>{doc.slug}</span>
								</CardHeader>
							</Card>
						</Link>
					</StaggerItem>
				))}
			</div>
		)
	})()

	return (
		<div className='space-y-4'>
			<DocsFilters />
			{list}
		</div>
	)
}

export { Docs }

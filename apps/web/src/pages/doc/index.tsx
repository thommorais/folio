import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { useSlugSync } from '_/routing/use-slug-sync'
import { RecordGone } from '_/components/record/record-gone'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Markdown } from '_/components/markdown'
import { useEntry } from '_/app/use-entry'
import type { Entry } from '_/core/domain/entry'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const DocBody = ({ doc }: { readonly doc: Entry }) => (
	<article className='space-y-8'>
		<header className='space-y-3'>
			<Heading>{doc.title}</Heading>

			<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
				<span className='font-mono'>{doc.slug}</span>
				<span>Updated {formatDate(doc.updatedAt)}</span>
				{doc.tags.map(tag => (
					<Badge key={tag} color='muted'>
						{tag}
					</Badge>
				))}
			</div>
		</header>

		{doc.body ? <Markdown>{doc.body}</Markdown> : <p className='text-dim text-sm'>This doc is empty.</p>}
	</article>
)

const DocDetail = () => {
	const { slug, doc } = useParams({ from: '/_authenticated/$slug/docs/$doc' })
	const state = useEntry(slug, doc)

	const navigate = useNavigate()

	useSlugSync({
		current: doc,
		record: state.status === 'ready' ? state.entry : undefined,
		rename: renamed => void navigate({ to: '/$slug/docs/$doc', params: { slug, doc: renamed }, replace: true }),
	})

	if (state.status === 'idle' || state.status === 'loading') {
		return <div className='bg-accent/40 h-32 animate-pulse' />
	}

	if (state.status === 'gone') {
		return (
			<RecordGone title={state.title}>
				<Link to='/$slug/docs' params={{ slug }} className='text-sm underline'>
					Back to docs
				</Link>
			</RecordGone>
		)
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return <DocBody doc={state.entry} />
}

export { DocDetail }

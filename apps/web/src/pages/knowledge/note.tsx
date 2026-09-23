import { Link, useParams } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Status } from '_/lib/async-status'
import { Markdown } from '_/components/markdown'
import { useKnowledge } from '_/app/use-knowledge'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

export const KnowledgeNote = () => {
	const { note: ref } = useParams({ from: '/_authenticated/knowledge/$note' })
	const state = useKnowledge(ref)

	if (state.status === Status.Loading) return <p className='text-dim text-sm'>Loading…</p>
	if (state.status === Status.Failed) return <p className='text-sm text-red-500'>{state.message}</p>

	const { note } = state

	return (
		<article className='space-y-8'>
			<header className='space-y-3'>
				<Link to='/knowledge' className='text-dimmer hover:text-accent text-xs'>
					← Knowledge
				</Link>

				<Heading>{note.title}</Heading>

				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<span className='font-mono'>{note.slug}</span>
					<span>{formatDate(note.updatedAt)}</span>
					{note.tags.map(tag => (
						<Badge key={tag} color='muted'>
							{tag}
						</Badge>
					))}
				</div>
			</header>

			{note.body ? <Markdown>{note.body}</Markdown> : <p className='text-dim text-sm'>This note has no body.</p>}
		</article>
	)
}

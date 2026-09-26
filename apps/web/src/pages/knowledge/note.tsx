import { Link, useParams } from '@tanstack/react-router'
import { Tag } from '_/components/issue/tag'
import { Heading } from '@thom/ui/heading'
import { Status } from '_/lib/async-status'
import { Markdown } from '_/components/markdown'
import { RecordGone } from '_/components/record/record-gone'
import { useKnowledge } from '_/app/use-knowledge'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

export const KnowledgeNote = () => {
	const { note: ref } = useParams({ from: '/_authenticated/knowledge/$note' })
	const state = useKnowledge(ref)

	if (state.status === Status.Idle || state.status === Status.Loading) return null
	if (state.status === Status.Failed) return <p className='text-destructive text-sm'>{state.message}</p>
	if (state.status === Status.Gone) {
		return (
			<RecordGone title={state.title}>
				<Link to='/knowledge' className='text-sm underline'>
					Back to knowledge
				</Link>
			</RecordGone>
		)
	}

	const { note } = state

	return (
		<article className='space-y-8'>
			<header className='space-y-3'>
				<Link to='/knowledge' className='text-dimmer hover:text-foreground text-xs transition-colors'>
					← Knowledge
				</Link>

				<Heading>{note.title}</Heading>

				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<span className='font-mono'>{note.slug}</span>
					<span>{formatDate(note.updatedAt)}</span>
					{note.tags.map(tag => (
						<Tag key={tag} tag={tag} />
					))}
				</div>
			</header>

			{note.body ? <Markdown>{note.body}</Markdown> : <p className='text-dim text-sm'>This note has no body.</p>}
		</article>
	)
}

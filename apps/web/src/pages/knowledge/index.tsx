import { Link } from '@tanstack/react-router'
import { useState } from 'react'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Status } from '_/lib/async-status'
import { useKnowledgeList } from '_/app/use-knowledge'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

export const KnowledgeList = () => {
	const [term, setTerm] = useState('')
	const state = useKnowledgeList(term)

	return (
		<section className='space-y-6'>
			<header className='space-y-2'>
				<Heading>Knowledge</Heading>
				<p className='text-dim text-sm'>
					Tips, snippets and fixes worth keeping. Not scoped to a project, so anything here is findable from anywhere.
				</p>
			</header>

			<input
				className='border-line bg-surface w-full rounded-md border px-3 py-2 text-sm'
				placeholder='Filter by title or body'
				value={term}
				onChange={event => setTerm(event.target.value)}
			/>

			{state.status === Status.Loading && <p className='text-dim text-sm'>Loading…</p>}
			{state.status === Status.Failed && <p className='text-sm text-red-500'>{state.message}</p>}

			{state.status === Status.Ready && state.notes.length === 0 && (
				<p className='text-dim text-sm'>Nothing here yet. Add one with `folio kb add`.</p>
			)}

			{state.status === Status.Ready && state.notes.length > 0 && (
				<ul className='divide-line divide-y'>
					{state.notes.map(note => (
						<li key={note.id} className='py-3'>
							<Link
								to='/knowledge/$note'
								params={{ note: note.slug }}
								className='hover:text-accent flex flex-col gap-1'
							>
								<span className='font-medium'>{note.title}</span>
								<span className='text-dimmer flex flex-wrap items-center gap-2 text-xs'>
									<span className='font-mono'>{note.slug}</span>
									<span>{formatDate(note.updatedAt)}</span>
									{note.tags.map(tag => (
										<Badge key={tag} color='muted'>
											{tag}
										</Badge>
									))}
								</span>
							</Link>
						</li>
					))}
				</ul>
			)}
		</section>
	)
}

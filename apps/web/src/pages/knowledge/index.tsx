import { Link, useNavigate, useSearch } from '@tanstack/react-router'
import { Heading } from '@thom/ui/heading'
import { useKnowledgeList } from '_/app/use-knowledge'
import { Tags } from '_/components/issue/tag'
import { FilterBar } from '_/components/list/filter-bar'
import { EmptyState } from '_/components/empty-state'
import { StaggerItem } from '_/components/motion/stagger'
import { Flash } from '_/components/motion/flash'
import { Status } from '_/lib/async-status'
import type { KnowledgeSearch } from '_/routes/_authenticated/knowledge'

const dayMonth = new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short' })

export const KnowledgeList = () => {
	const search = useSearch({ from: '/_authenticated/knowledge/' })
	const navigate = useNavigate()
	const state = useKnowledgeList(search.q ?? '')

	const list = (() => {
		if (state.status === Status.Loading) return null

		if (state.status === Status.Failed) {
			return <p className='text-destructive text-sm'>{state.message}</p>
		}

		if (state.notes.length === 0) {
			if (search.q !== undefined) return <p className='text-dim text-sm'>No notes match.</p>

			return <EmptyState message='No notes yet.' command={'folio kb add "<title>"'} />
		}

		return (
			<div className='border-border divide-border divide-y border'>
				{state.notes.map((note, index) => (
					<StaggerItem key={note.id} index={index}>
						<Link
							to='/knowledge/$note'
							params={{ note: note.slug }}
							className='hover:bg-accent/40 active:bg-accent/60 relative flex h-11 items-center gap-3 px-4 transition-colors'
						>
							<Flash on={note.updatedAt.getTime()} />
							<span className='min-w-0 flex-1 truncate text-sm'>{note.title}</span>
							<Tags tags={note.tags} className='hidden shrink-0 flex-nowrap sm:flex' />
							<span className='text-dimmer w-12 shrink-0 text-right text-xs tabular-nums'>
								{dayMonth.format(note.updatedAt)}
							</span>
						</Link>
					</StaggerItem>
				))}
			</div>
		)
	})()

	return (
		<section className='space-y-6'>
			<header className='space-y-2'>
				<Heading>Knowledge</Heading>
				<p className='text-dim text-sm'>
					Tips, snippets and fixes worth keeping. Not scoped to a project, so anything here is findable from anywhere.
				</p>
			</header>

			<div className='space-y-4'>
				<FilterBar
					placeholder='Search knowledge...'
					term={search.q}
					chips={[]}
					onSearch={q => {
						void navigate({ from: '/knowledge/', to: '.', search: (prev: KnowledgeSearch) => ({ ...prev, q }) })
					}}
				/>
				{list}
			</div>
		</section>
	)
}

import { createFileRoute } from '@tanstack/react-router'
import { KnowledgeNote } from '_/pages/knowledge/note'

export const Route = createFileRoute('/_authenticated/knowledge/$note')({
	component: () => (
		<div className='mx-auto w-full max-w-5xl'>
			<KnowledgeNote />
		</div>
	),
})

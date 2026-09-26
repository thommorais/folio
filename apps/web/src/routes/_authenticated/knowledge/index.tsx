import { createFileRoute } from '@tanstack/react-router'
import { KnowledgeList } from '_/pages/knowledge'
import { asString } from '_/routes/search-params'

export type KnowledgeSearch = {
	readonly q?: string
}

export const Route = createFileRoute('/_authenticated/knowledge/')({
	validateSearch: (search: Record<string, unknown>): KnowledgeSearch => ({
		q: asString(search.q),
	}),
	component: () => (
		<div className='mx-auto w-full max-w-5xl'>
			<KnowledgeList />
		</div>
	),
})

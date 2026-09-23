import { createFileRoute } from '@tanstack/react-router'
import { KnowledgeList } from '_/pages/knowledge'

export const Route = createFileRoute('/_authenticated/knowledge/')({
	component: KnowledgeList,
})

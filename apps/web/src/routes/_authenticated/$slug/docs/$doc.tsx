import { createFileRoute } from '@tanstack/react-router'
import { DocDetail } from '_/pages/doc'

export const Route = createFileRoute('/_authenticated/$slug/docs/$doc')({
	component: DocDetail,
})

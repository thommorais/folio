import { createFileRoute } from '@tanstack/react-router'
import { SharedPage } from '_/pages/share'

export const Route = createFileRoute('/share/$token')({
	component: SharedPage,
})

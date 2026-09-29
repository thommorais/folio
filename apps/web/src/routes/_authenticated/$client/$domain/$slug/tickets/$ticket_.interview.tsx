import { createFileRoute } from '@tanstack/react-router'
import { InterviewPage } from '_/pages/interview'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/tickets/$ticket_/interview')({
	component: InterviewPage,
})

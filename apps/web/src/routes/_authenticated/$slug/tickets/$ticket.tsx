import { createFileRoute } from '@tanstack/react-router'
import { TicketDetail } from '_/pages/ticket'

export const Route = createFileRoute('/_authenticated/$slug/tickets/$ticket')({
	component: TicketDetail,
})

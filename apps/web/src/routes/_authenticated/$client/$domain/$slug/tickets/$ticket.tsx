import { createFileRoute } from '@tanstack/react-router'
import { TicketDetail } from '_/pages/ticket'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/tickets/$ticket')({
	component: TicketDetail,
})

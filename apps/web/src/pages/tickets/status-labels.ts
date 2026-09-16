import type { TicketStatus } from '_/core/domain/ticket'

export const TICKET_STATUS_LABELS: Record<TicketStatus, string> = {
	open: 'Open',
	in_progress: 'In progress',
	blocked: 'Blocked',
	closed: 'Closed',
	cancelled: 'Cancelled',
}

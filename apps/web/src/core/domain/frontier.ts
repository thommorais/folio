import type { Ticket, TicketId } from './ticket'
import { isTerminal } from './ticket'

export type Partitioned = {
	readonly frontier: readonly Ticket[]
	readonly blocked: readonly Ticket[]
	readonly claimed: readonly Ticket[]
	readonly done: readonly Ticket[]
}

const oldestFirst = (a: Ticket, b: Ticket): number => a.createdAt.getTime() - b.createdAt.getTime()

export const isBlocked = (ticket: Ticket, byId: ReadonlyMap<TicketId, Ticket>): boolean =>
	ticket.dependsOn.some(id => {
		const blocker = byId.get(id)

		return blocker !== undefined && !isTerminal(blocker.status)
	})

export const partitionChildren = (children: readonly Ticket[]): Partitioned => {
	const byId = new Map(children.map(child => [child.id, child]))

	const frontier: Ticket[] = []
	const blocked: Ticket[] = []
	const claimed: Ticket[] = []
	const done: Ticket[] = []

	for (const child of children) {
		if (isTerminal(child.status)) {
			done.push(child)
		} else if (isBlocked(child, byId)) {
			blocked.push(child)
		} else if (child.assignee !== undefined) {
			claimed.push(child)
		} else {
			frontier.push(child)
		}
	}

	return {
		frontier: frontier.sort(oldestFirst),
		blocked: blocked.sort(oldestFirst),
		claimed: claimed.sort(oldestFirst),
		done: done.sort(oldestFirst),
	}
}

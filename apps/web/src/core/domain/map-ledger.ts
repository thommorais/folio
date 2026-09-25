import type { Cycle } from './cycle'
import { isResolved, PHASE } from './cycle'
import type { Issue } from './issue'
import { ISSUE_KIND, ISSUE_STATUS, isTerminal } from './issue'

export type Ledger = {
	readonly decided: readonly Issue[]
	readonly outOfScope: readonly Issue[]
}

const oldestFirst = (a: Issue, b: Issue): number => a.createdAt.getTime() - b.createdAt.getTime()

const decisions = (children: readonly Issue[]): readonly Issue[] => children.filter(child => child.kind === ISSUE_KIND.TICKET)

export const mapLedger = (children: readonly Issue[]): Ledger => {
	const tickets = decisions(children)

	return {
		decided: tickets.filter(child => child.status === ISSUE_STATUS.DONE).sort(oldestFirst),
		outOfScope: tickets.filter(child => child.status === ISSUE_STATUS.CANCELLED).sort(oldestFirst),
	}
}

export const planHold = (cycle: Cycle, mapChildren: readonly Issue[]): number => {
	if (cycle.mapId === undefined || cycle.phase !== PHASE.PLAN || isResolved(cycle)) return 0

	return decisions(mapChildren).filter(child => !isTerminal(child.status)).length
}

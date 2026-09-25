import type { Branded } from './branded'
import type { ProjectId, UserId } from './project'
import type { IssueId } from './issue'

export type CycleId = Branded<string, 'CycleId'>

export const cycleId = (value: string): CycleId => value as CycleId

export const PHASE = {
	PLAN: 'plan',
	DO: 'do',
	CHECK: 'check',
	ACT: 'act',
} as const

export const PHASES = [PHASE.PLAN, PHASE.DO, PHASE.CHECK, PHASE.ACT] as const

export type Phase = (typeof PHASES)[number]

export type Cycle = {
	readonly id: CycleId
	readonly projectId: ProjectId
	readonly ticketId: IssueId
	readonly ordinal: number
	readonly phase: Phase
	readonly resolution: string
	readonly mapId: IssueId | undefined
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
	readonly closedAt: Date | undefined
}

export const isResolved = (cycle: Cycle): boolean => cycle.closedAt !== undefined

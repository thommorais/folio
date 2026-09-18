import type { Branded } from './branded'
import type { ProjectId, UserId } from './project'
import type { IssueId } from './issue'

export type CycleId = Branded<string, 'CycleId'>

export const cycleId = (value: string): CycleId => value as CycleId

export const PHASES = ['plan', 'do', 'check', 'act'] as const

export type Phase = (typeof PHASES)[number]

export type Cycle = {
	readonly id: CycleId
	readonly projectId: ProjectId
	readonly ticketId: IssueId
	readonly ordinal: number
	readonly phase: Phase
	readonly resolution: string
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
	readonly closedAt: Date | undefined
}

export const isResolved = (cycle: Cycle): boolean => cycle.closedAt !== undefined

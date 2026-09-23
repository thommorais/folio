import type { Branded } from './branded';
import type { IssueId } from './issue';
import type { ProjectId, UserId } from './project';

export type PlanId = Branded<string, 'PlanId'>

export const planId = (value: string): PlanId => value as PlanId

const STATUS = {
	DRAFT:  'draft',
	ACTIVE:'active',
	DONE: 'done',
	ABANDONED: 'abandoned'
} as const

export const PLAN_STATUSES = [STATUS.DRAFT, STATUS.ACTIVE, STATUS.DONE, STATUS.ABANDONED] as const

export type PlanStatus = (typeof PLAN_STATUSES)[number]

export const DEFAULT_PLAN_STATUSES: readonly PlanStatus[] = PLAN_STATUSES.filter(status => status !== STATUS.DONE && status !== STATUS.ABANDONED )

export type Plan = {
	readonly id: PlanId
	readonly projectId: ProjectId
	readonly ticketId: IssueId | undefined
	readonly title: string
	readonly goal: string
	readonly status: PlanStatus
	readonly tags: readonly string[]
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

export const isTerminal = (status: PlanStatus): boolean => status === 'done' || status === 'abandoned'

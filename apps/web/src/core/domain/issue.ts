import type { Branded } from './branded'
import type { PlanId } from './plan'
import type { ProjectId, UserId } from './project'

export type IssueId = Branded<string, 'IssueId'>

export const issueId = (value: string): IssueId => value as IssueId

export const ISSUE_KINDS = ['ticket', 'todo'] as const

export type IssueKind = (typeof ISSUE_KINDS)[number]

export const ISSUE_STATUSES = ['open', 'in_progress', 'blocked', 'done', 'cancelled'] as const

export type IssueStatus = (typeof ISSUE_STATUSES)[number]

export const DEFAULT_TICKET_STATUSES: readonly IssueStatus[] = ISSUE_STATUSES.filter(status => status !== 'done')

export const PRIORITIES = ['low', 'medium', 'high'] as const

export type Priority = (typeof PRIORITIES)[number]

export const SIZES = [1, 2, 3, 5, 8] as const

export type Size = (typeof SIZES)[number]

export const WAYFINDER_TYPES = ['map', 'research', 'prototype', 'grilling', 'task'] as const

export type WayfinderType = (typeof WAYFINDER_TYPES)[number]

export const LINK_KINDS = ['blocks', 'relates', 'parent'] as const

export type LinkKind = (typeof LINK_KINDS)[number]

export type IssueLink = {
	readonly from: IssueId
	readonly to: IssueId
	readonly kind: LinkKind
}

export type Issue = {
	readonly id: IssueId
	readonly kind: IssueKind
	readonly projectId: ProjectId
	readonly planId: PlanId | undefined
	readonly slug: string
	readonly title: string
	readonly body: string
	readonly status: IssueStatus
	readonly priority: Priority
	readonly size: Size | undefined
	readonly assignee: UserId | undefined
	readonly tags: readonly string[]
	readonly position: number
	readonly dueDate: Date | undefined
	readonly wayfinder: WayfinderType | undefined
	readonly externalRef: string
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
	readonly parentId: IssueId | undefined
	readonly dependsOn: readonly IssueId[]
	readonly relatedTo: readonly IssueId[]
	readonly blocked: boolean
}

export const isTerminal = (status: IssueStatus): boolean => status === 'done' || status === 'cancelled'

const PRIORITY_WEIGHT: Record<Priority, number> = { low: 1, medium: 2, high: 3 }

export const score = (issue: Pick<Issue, 'priority' | 'size'>): number =>
	issue.size === undefined ? 0 : PRIORITY_WEIGHT[issue.priority] / issue.size

export const isSize = (value: number): value is Size => (SIZES as readonly number[]).includes(value)

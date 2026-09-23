import type { Branded } from './branded'
import type { EntryKind } from './entry'
import type { IssueKind, IssueStatus, Priority } from './issue'
import type { Phase } from './cycle'
import type { PlanStatus } from './plan'
import type { ProjectId } from './project'

export type SharedIssue = {
	readonly kind: IssueKind
	readonly title: string
	readonly body: string
	readonly status: IssueStatus
	readonly priority: Priority
	readonly tags: readonly string[]
	readonly externalRef: string
	readonly updatedAt: Date
}

export type SharedPlan = {
	readonly title: string
	readonly goal: string
	readonly status: PlanStatus
	readonly tags: readonly string[]
	readonly progress: { readonly total: number; readonly done: number }
	readonly updatedAt: Date
}

export type SharedEntry = {
	readonly id: string
	readonly kind: EntryKind
	readonly title: string
	readonly body: string
	readonly createdAt: Date
}

export type SharedCycle = {
	readonly id: string
	readonly ordinal: number
	readonly phase: Phase
	readonly resolution: string
	readonly closedAt: Date | undefined
}

export type SharedItem =
	| {
			readonly kind: 'issue'
			readonly label: string
			readonly issue: SharedIssue
			readonly todos: readonly (SharedIssue & { readonly id: string })[]
			readonly plans: readonly (SharedPlan & { readonly id: string })[]
			readonly journal: readonly SharedEntry[]
			readonly docs: readonly SharedEntry[]
			readonly cycles: readonly SharedCycle[]
	  }
	| {
			readonly kind: 'plan'
			readonly label: string
			readonly plan: SharedPlan
			readonly todos: readonly (SharedIssue & { readonly id: string })[]
	  }

export type ShareId = Branded<string, 'ShareId'>

export const shareId = (value: string): ShareId => value as ShareId

export type ShareTarget = {
	readonly kind: 'issue' | 'plan'
	readonly id: string
	readonly projectId: ProjectId
}

export type ShareLink = {
	readonly id: ShareId
	readonly label: string
	readonly token: string
	readonly createdAt: Date
	readonly lastAccessedAt: Date | undefined
}

export const SHARE_LABEL_MAX = 120

export const shareUrl = (origin: string, token: string): string => `${origin}/share/${token}`

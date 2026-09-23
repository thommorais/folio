import type { Branded } from './branded';
import type { CycleId } from './cycle';
import type { IssueId } from './issue';
import type { PlanId } from './plan';
import type { ProjectId, UserId } from './project';

export type EntryId = Branded<string, 'EntryId'>

export const entryId = (value: string): EntryId => value as EntryId

export const ENTRY_KINDS = ['journal', 'doc', 'log'] as const

export type EntryKind = (typeof ENTRY_KINDS)[number]

export type Entry = {
	readonly id: EntryId
	readonly kind: EntryKind
	readonly projectId: ProjectId
	readonly issueId: IssueId | undefined
	readonly planId: PlanId | undefined
	readonly cycleId: CycleId | undefined
	readonly slug: string
	readonly title: string
	readonly body: string
	readonly branch: string
	readonly pr: string
	readonly externalRef: string
	readonly tags: readonly string[]
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

// Journal entries and docs are addressed by slug and listed on their own.
// A log belongs to the ticket it was written against, so it is not one of them.
export const ADDRESSABLE_KINDS = ['journal', 'doc'] as const

export type AddressableKind = (typeof ADDRESSABLE_KINDS)[number]

export const isAddressable = (kind: EntryKind): boolean => kind === 'journal' || kind === 'doc'

export const KIND = {
	TICKET: 'ticket',
	TAGS: 'tags',
	TODO: 'todo',
	LOG: 'log'
} as const

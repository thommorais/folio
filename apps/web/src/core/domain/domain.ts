import type { Branded } from './branded'
import type { ClientId } from './client'
import type { UserId } from './project'

export type DomainId = Branded<string, 'DomainId'>

export const domainId = (value: string): DomainId => value as DomainId

export const ROLE = {
	OWNER: 'owner',
	EDITOR: 'editor',
	VIEWER: 'viewer',
} as const

export const ROLES = [ROLE.OWNER, ROLE.EDITOR, ROLE.VIEWER] as const

export type Role = (typeof ROLES)[number]

export type Member = {
	readonly userId: UserId
	readonly role: Role
	readonly email: string
	readonly name: string
}

// Membership hangs off the domain, not the project: one roster covers every
// project a client's domain holds, which is what the PocketBase rules check.
export type Domain = {
	readonly id: DomainId
	readonly clientId: ClientId
	readonly slug: string
	readonly name: string
	readonly descr: string
	readonly members: readonly Member[]
	readonly createdAt: Date
	readonly updatedAt: Date
}

export const canWrite = (role: Role): boolean => role === ROLE.OWNER || role === ROLE.EDITOR

export const canAdmin = (role: Role): boolean => role === ROLE.OWNER

export const roleOf = (domain: Domain, user: UserId): Role | undefined =>
	domain.members.find(member => member.userId === user)?.role

export const ownersOf = (domain: Domain): readonly Member[] => domain.members.filter(member => member.role === ROLE.OWNER)

import type { Branded } from './branded'
export type ProjectId = Branded<string, 'ProjectId'>
export type UserId = Branded<string, 'UserId'>

export const projectId = (value: string): ProjectId => value as ProjectId
export const userId = (value: string): UserId => value as UserId

export const ROLES = ['owner', 'editor', 'viewer'] as const

export type Role = (typeof ROLES)[number]

export type Member = {
	readonly userId: UserId
	readonly role: Role
	readonly email: string
	readonly name: string
}

export type Project = {
	readonly id: ProjectId
	readonly slug: string
	readonly name: string
	readonly descr: string
	readonly archived: boolean
	readonly members: readonly Member[]
	readonly createdAt: Date
	readonly updatedAt: Date
}

export const canWrite = (role: Role): boolean => role === 'owner' || role === 'editor'

export const canAdmin = (role: Role): boolean => role === 'owner'

export const roleOf = (project: Project, user: UserId): Role | undefined =>
	project.members.find(member => member.userId === user)?.role

export const ownersOf = (project: Project): readonly Member[] =>
	project.members.filter(member => member.role === 'owner')

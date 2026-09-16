import type { Branded } from './branded'
import type { ProjectId, UserId } from './project'
import type { TicketId } from './ticket'

export type DocId = Branded<string, 'DocId'>

export const docId = (value: string): DocId => value as DocId

export type Doc = {
	readonly id: DocId
	readonly projectId: ProjectId
	readonly ticketId: TicketId | undefined
	readonly slug: string
	readonly title: string
	readonly body: string
	readonly tags: readonly string[]
	readonly createdBy: UserId | undefined
	readonly createdAt: Date
	readonly updatedAt: Date
}

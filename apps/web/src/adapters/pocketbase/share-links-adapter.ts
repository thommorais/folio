import type { ShareLink, ShareTarget } from '_/core/domain/share'
import { shareId as toShareId } from '_/core/domain/share'
import type { ShareLinksPort } from '_/core/ports/share-links'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournSharesResponse } from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { keyed } from './request-key'

const toShareLink = (record: JournSharesResponse): ShareLink => ({
	id: toShareId(record.id),
	label: record.label,
	token: record.token,
	createdAt: new Date(record.created),
	lastAccessedAt: record.last_accessed_at ? new Date(record.last_accessed_at) : undefined,
})

export const createShareLinksAdapter = (): ShareLinksPort => {
	const client = getPocketBaseClient()
	const collection = client.collection(Collections.JournShares)

	return {
		list: async (target: ShareTarget): Promise<Result<ReadonlyArray<ShareLink>>> => {
			const filter = client.filter(`${target.kind} = {:id}`, { id: target.id })

			const { data, error } = await tryCatch(
				collection.getFullList<JournSharesResponse>(keyed('shares.list', { filter, sort: '-created' })),
			)

			return error
				? err(new Error(`Failed to load share links: ${error.message}`, { cause: error }))
				: ok(data.map(toShareLink))
		},

		create: async (target, label): Promise<Result<ShareLink>> => {
			const { data, error } = await tryCatch(
				collection.create<JournSharesResponse>({
					project: target.projectId,
					[target.kind]: target.id,
					label: label.trim(),
					created_by: client.authStore.record?.id,
				}),
			)

			return error
				? err(new Error(`Could not create the link: ${error.message}`, { cause: error }))
				: ok(toShareLink(data))
		},

		revoke: async (id): Promise<Result<void>> => {
			const { error } = await tryCatch(collection.delete(id))

			return error ? err(new Error(`Could not revoke the link: ${error.message}`, { cause: error })) : ok(undefined)
		},
	}
}

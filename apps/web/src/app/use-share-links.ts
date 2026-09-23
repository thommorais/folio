import { useCallback } from 'react'
import type { ShareId, ShareLink, ShareTarget } from '_/core/domain/share'
import type { Result } from '_/lib/result'
import { useAsyncState } from './realtime/use-async-state'
import { useContainer } from './container'

export const useShareLinks = (target: ShareTarget, skip = false) => {
	const { shareLinks } = useContainer()

	const { state, refetch } = useAsyncState(() => shareLinks.list(target), [target.kind, target.id, shareLinks], skip)

	const create = useCallback(
		async (label: string): Promise<Result<ShareLink>> => {
			const result = await shareLinks.create(target, label)
			if (result.success) await refetch()
			return result
		},
		[shareLinks, target, refetch],
	)

	const revoke = useCallback(
		async (id: ShareId): Promise<Result<void>> => {
			const result = await shareLinks.revoke(id)
			if (result.success) await refetch()
			return result
		},
		[shareLinks, refetch],
	)

	return { state, create, revoke }
}

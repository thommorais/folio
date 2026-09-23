import type { Result } from '_/lib/result'
import type { ShareId, ShareLink, ShareTarget } from '../domain/share'

export type ShareLinksPort = {
	readonly list: (target: ShareTarget) => Promise<Result<ReadonlyArray<ShareLink>>>
	readonly create: (target: ShareTarget, label: string) => Promise<Result<ShareLink>>
	readonly revoke: (id: ShareId) => Promise<Result<void>>
}

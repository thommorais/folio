import type { Result } from '_/lib/result'
import type { SharedItem } from '../domain/share'

export type SharesPort = {
	readonly open: (token: string) => Promise<Result<SharedItem | undefined>>
}

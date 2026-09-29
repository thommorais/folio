import type { Result } from '_/lib/result'
import type { Interview, SendAction, Sent } from '../domain/interview'
import type { Unsubscribe } from './subscription'

export type InterviewsPort = {
	readonly current: (issueId: string) => Promise<Result<Interview>>
	readonly subscribeToRecord: (
		id: string,
		onChange: (interview: Interview) => void,
		onGone: () => void,
	) => Promise<Result<Unsubscribe>>
	readonly send: (issueId: string, actions: readonly SendAction[]) => Promise<Result<Sent>>
}

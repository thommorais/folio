import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'
import type { Interview } from '_/core/domain/interview'
import { Status } from '_/lib/async-status'

type InterviewState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly interview: Interview }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useInterview = (issueId: string): InterviewState => {
	const { interviews, connection } = useContainer()

	const state = useLiveRecord<Interview>({
		load: () => interviews.current(issueId),
		subscribe: (id, onChange, onGone) => interviews.subscribeToRecord(id, onChange, onGone),
		titleOf: interview => interview.topic,
		connection,
		deps: [issueId, interviews],
		skip: !issueId,
	})

	return state.status === Status.Ready ? { status: Status.Ready, interview: state.data } : state
}

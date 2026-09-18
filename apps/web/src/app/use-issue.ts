import type { Issue } from '_/core/domain/issue'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'

type IssueState =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly issue: Issue }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

export const useIssue = (project: string, slug: string): IssueState => {
	const { issues, connection } = useContainer()

	const state = useLiveRecord<Issue>({
		load: () => issues.get(project, slug),
		subscribe: (id, onChange, onGone) => issues.subscribeToRecord(project, id, onChange, onGone),
		connection,
		deps: [project, slug],
		skip: !slug,
	})

	return state.status === 'ready' ? { status: 'ready', issue: state.data } : state
}

export const useIssueById = (project: string, id: string): IssueState => {
	const { issues, connection } = useContainer()

	const state = useLiveRecord<Issue>({
		load: () => issues.getById(project, id),
		subscribe: (recordId, onChange, onGone) => issues.subscribeToRecord(project, recordId, onChange, onGone),
		connection,
		deps: [project, id],
		skip: !id,
	})

	return state.status === 'ready' ? { status: 'ready', issue: state.data } : state
}

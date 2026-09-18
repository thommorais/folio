import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Issue } from '_/core/domain/issue'
import type { IssueFilter } from '_/core/ports/issues'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type IssuesState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly issues: readonly Issue[] }
	| { readonly status: 'failed'; readonly message: string }

export const useIssues = (project: string, filter?: IssueFilter): IssuesState => {
	const { issues, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Issue>({
		load: () => issues.list(project, JSON.parse(key) as IssueFilter),
		subscribe: update => issues.subscribeToList(project, update, JSON.parse(key) as IssueFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, issues],
	})

	return state.status === 'ready' ? { status: 'ready', issues: state.data } : state
}

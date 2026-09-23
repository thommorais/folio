import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import { withBlocked } from '_/core/domain/blocked'
import type { Issue } from '_/core/domain/issue'
import type { IssueFilter } from '_/core/ports/issues'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type IssuesState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly issues: readonly Issue[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useIssues = (project: string, filter?: IssueFilter): IssuesState => {
	const { issues, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Issue>({
		load: () => issues.list(project, JSON.parse(key) as IssueFilter),
		subscribe: update => issues.subscribeToList(project, update, JSON.parse(key) as IssueFilter),
		// An event carries the issue that changed, never the ones whose blocked
		// flag it invalidates, so the whole set is rederived after each fold.
		fold: (rows, row, action) => withBlocked(foldUpdates(rows, row, action)),
		connection,
		deps: [project, key, issues],
	})

	return state.status === Status.Ready ? { status: Status.Ready, issues: state.data } : state
}

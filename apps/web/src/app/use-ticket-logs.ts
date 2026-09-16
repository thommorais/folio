import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { TicketLog } from '_/core/domain/worklog'
import type { TicketLogFilter } from '_/core/ports/worklogs'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type TicketLogsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly logs: readonly TicketLog[] }
	| { readonly status: 'failed'; readonly message: string }

export const useTicketLogs = (project: string, filter?: TicketLogFilter): TicketLogsState => {
	const { workLogs, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<TicketLog>({
		load: () => workLogs.listTicketLogs(project, JSON.parse(key) as TicketLogFilter),
		subscribe: update => workLogs.subscribeToTicketLogs(project, update, JSON.parse(key) as TicketLogFilter),
		fold: foldUpdates,
		connection,
		deps: [project, key, workLogs],
	})

	return state.status === 'ready' ? { status: 'ready', logs: state.data } : state
}

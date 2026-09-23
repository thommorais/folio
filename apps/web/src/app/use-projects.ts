import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Project } from '_/core/domain/project'
import type { ProjectFilter } from '_/core/ports/projects'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type ProjectsState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly projects: readonly Project[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useProjects = (filter?: ProjectFilter, skip = false): ProjectsState => {
	const { projects, connection } = useContainer()
	const key = useFilterKey(filter)

	const state = useLiveList<Project>({
		load: () => projects.list(JSON.parse(key) as ProjectFilter),
		subscribe: update => projects.subscribeToList(update, JSON.parse(key) as ProjectFilter),
		fold: foldUpdates,
		connection,
		deps: [key, projects],
		skip,
	})

	return state.status === Status.Ready ? { status: Status.Ready, projects: state.data } : state
}

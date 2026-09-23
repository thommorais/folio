import { foldUpdates } from '_/adapters/pocketbase/fold-updates'
import type { Project } from '_/core/domain/project'
import type { ProjectFilter } from '_/core/ports/projects'
import { useFilterKey } from './realtime/use-filter-key'
import { useLiveList } from './realtime/use-live-list'
import { useContainer } from './container'

type ProjectsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly projects: readonly Project[] }
	| { readonly status: 'failed'; readonly message: string }

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

	return state.status === 'ready' ? { status: 'ready', projects: state.data } : state
}

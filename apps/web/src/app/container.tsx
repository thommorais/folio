import { createContext, useContext, useMemo } from 'react'
import { createThemeAdapter } from '_/adapters/browser/theme-adapter'
import { createAuthAdapter } from '_/adapters/pocketbase/auth-adapter'
import { createConnectionAdapter } from '_/adapters/pocketbase/connection-adapter'
import { createCyclesAdapter } from '_/adapters/pocketbase/cycles-adapter'
import { createEntriesAdapter } from '_/adapters/pocketbase/entries-adapter'
import { createIssuesAdapter } from '_/adapters/pocketbase/issues-adapter'
import { createPlansAdapter } from '_/adapters/pocketbase/plans-adapter'
import { createProjectsAdapter } from '_/adapters/pocketbase/projects-adapter'
import { createSearchAdapter } from '_/adapters/pocketbase/search-adapter'
import type { AuthPort } from '_/core/ports/auth'
import type { ConnectionPort } from '_/core/ports/connection'
import type { CyclesPort } from '_/core/ports/cycles'
import type { EntriesPort } from '_/core/ports/entries'
import type { IssuesPort } from '_/core/ports/issues'
import type { PlansPort } from '_/core/ports/plans'
import type { ProjectsPort } from '_/core/ports/projects'
import type { SearchPort } from '_/core/ports/search'
import type { ThemePort } from '_/core/ports/theme'

export type Container = {
	readonly auth: AuthPort
	readonly connection: ConnectionPort
	readonly cycles: CyclesPort
	readonly entries: EntriesPort
	readonly issues: IssuesPort
	readonly plans: PlansPort
	readonly projects: ProjectsPort
	readonly search: SearchPort
	readonly theme: ThemePort
}

export const createContainer = (): Container => ({
	auth: createAuthAdapter(),
	connection: createConnectionAdapter(),
	cycles: createCyclesAdapter(),
	entries: createEntriesAdapter(),
	issues: createIssuesAdapter(),
	plans: createPlansAdapter(),
	projects: createProjectsAdapter(),
	search: createSearchAdapter(),
	theme: createThemeAdapter(),
})

const ContainerContext = createContext<Container | undefined>(undefined)

type ProviderProps = {
	readonly container?: Container
	readonly children: React.ReactNode
}

export const ContainerProvider = ({ container, children }: ProviderProps) => {
	const value = useMemo(() => container ?? createContainer(), [container])

	return <ContainerContext.Provider value={value}>{children}</ContainerContext.Provider>
}

export const useContainer = (): Container => {
	const container = useContext(ContainerContext)

	if (!container) {
		throw new Error('useContainer must be used inside a ContainerProvider')
	}

	return container
}

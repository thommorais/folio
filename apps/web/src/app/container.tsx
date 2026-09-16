import { createContext, useContext, useMemo } from 'react'
import { createThemeAdapter } from '_/adapters/browser/theme-adapter'
import { createAuthAdapter } from '_/adapters/pocketbase/auth-adapter'
import { createConnectionAdapter } from '_/adapters/pocketbase/connection-adapter'
import { createCyclesAdapter } from '_/adapters/pocketbase/cycles-adapter'
import { createDocsAdapter } from '_/adapters/pocketbase/docs-adapter'
import { createJournalAdapter } from '_/adapters/pocketbase/journal-adapter'
import { createPlansAdapter } from '_/adapters/pocketbase/plans-adapter'
import { createProjectsAdapter } from '_/adapters/pocketbase/projects-adapter'
import { createSearchAdapter } from '_/adapters/pocketbase/search-adapter'
import { createTicketsAdapter } from '_/adapters/pocketbase/tickets-adapter'
import { createTodosAdapter } from '_/adapters/pocketbase/todos-adapter'
import { createWorkLogsAdapter } from '_/adapters/pocketbase/worklogs-adapter'
import type { AuthPort } from '_/core/ports/auth'
import type { ConnectionPort } from '_/core/ports/connection'
import type { CyclesPort } from '_/core/ports/cycles'
import type { DocsPort } from '_/core/ports/docs'
import type { JournalPort } from '_/core/ports/journal'
import type { PlansPort } from '_/core/ports/plans'
import type { ProjectsPort } from '_/core/ports/projects'
import type { SearchPort } from '_/core/ports/search'
import type { ThemePort } from '_/core/ports/theme'
import type { TicketsPort } from '_/core/ports/tickets'
import type { TodosPort } from '_/core/ports/todos'
import type { WorkLogsPort } from '_/core/ports/worklogs'

export type Container = {
	readonly auth: AuthPort
	readonly connection: ConnectionPort
	readonly cycles: CyclesPort
	readonly docs: DocsPort
	readonly journal: JournalPort
	readonly plans: PlansPort
	readonly projects: ProjectsPort
	readonly search: SearchPort
	readonly theme: ThemePort
	readonly tickets: TicketsPort
	readonly todos: TodosPort
	readonly workLogs: WorkLogsPort
}

export const createContainer = (): Container => ({
	auth: createAuthAdapter(),
	connection: createConnectionAdapter(),
	cycles: createCyclesAdapter(),
	docs: createDocsAdapter(),
	journal: createJournalAdapter(),
	plans: createPlansAdapter(),
	projects: createProjectsAdapter(),
	search: createSearchAdapter(),
	theme: createThemeAdapter(),
	tickets: createTicketsAdapter(),
	todos: createTodosAdapter(),
	workLogs: createWorkLogsAdapter(),
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

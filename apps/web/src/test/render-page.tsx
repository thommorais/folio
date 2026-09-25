import { cleanup, render } from '@testing-library/react'
import {
	createMemoryHistory,
	createRootRoute,
	createRoute,
	createRouter,
	Outlet,
	RouterProvider,
} from '@tanstack/react-router'
import { ContainerProvider, type Container } from '_/app/container'
import type { Cycle } from '_/core/domain/cycle'
import type { Entry } from '_/core/domain/entry'
import type { Issue } from '_/core/domain/issue'
import type { CycleFilter } from '_/core/ports/cycles'
import type { EntryFilter } from '_/core/ports/entries'
import type { IssueFilter } from '_/core/ports/issues'
import { ok } from '_/lib/result'
import { afterEach } from 'vitest'

afterEach(cleanup)

type Records = {
	readonly issues?: readonly Issue[]
	readonly entries?: readonly Entry[]
	readonly cycles?: readonly Cycle[]
}

const unsubscribed = async () => ok(async () => {})

const issuesMatching = (issues: readonly Issue[], filter: IssueFilter = {}) =>
	issues.filter(
		issue =>
			(filter.kind === undefined || issue.kind === filter.kind) &&
			(filter.parentId === undefined || issue.parentId === filter.parentId),
	)

const entriesMatching = (entries: readonly Entry[], filter: EntryFilter = {}) =>
	entries.filter(
		entry =>
			(filter.kind === undefined || entry.kind === filter.kind) &&
			(filter.kinds === undefined || filter.kinds.includes(entry.kind)) &&
			(filter.issueId === undefined || entry.issueId === filter.issueId) &&
			(filter.cycleId === undefined || entry.cycleId === filter.cycleId),
	)

const cyclesMatching = (cycles: readonly Cycle[], filter: CycleFilter = {}) =>
	cycles.filter(
		cycle =>
			(filter.ticketId === undefined || cycle.ticketId === filter.ticketId) &&
			(filter.mapId === undefined || cycle.mapId === filter.mapId),
	)

const containerOf = ({ issues = [], entries = [], cycles = [] }: Records): Container =>
	({
		connection: { onReconnect: () => () => {} },
		issues: {
			list: async (_project: string, filter?: IssueFilter) => ok(issuesMatching(issues, filter)),
			subscribeToList: unsubscribed,
		},
		entries: {
			list: async (_project: string, filter?: EntryFilter) => ok(entriesMatching(entries, filter)),
			subscribeToList: unsubscribed,
		},
		cycles: {
			list: async (_project: string, filter?: CycleFilter) => ok(cyclesMatching(cycles, filter)),
			subscribeToList: unsubscribed,
		},
		plans: {
			list: async () => ok([]),
			subscribeToList: unsubscribed,
		},
	}) as unknown as Container

export const renderPage = (ui: React.ReactNode, records: Records = {}) => {
	const root = createRootRoute({ component: Outlet })
	const page = createRoute({ getParentRoute: () => root, path: '/$client/$domain/$slug', component: () => ui })
	const router = createRouter({
		routeTree: root.addChildren([page]),
		history: createMemoryHistory({ initialEntries: ['/acme/web/folio'] }),
	})

	return render(
		<ContainerProvider container={containerOf(records)}>
			<RouterProvider router={router} />
		</ContainerProvider>,
	)
}

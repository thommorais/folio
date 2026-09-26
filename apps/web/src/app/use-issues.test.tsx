import { act, renderHook, waitFor } from '@testing-library/react'
import type { Issue } from '_/core/domain/issue'
import type { ConnectionPort } from '_/core/ports/connection'
import type { IssuesPort } from '_/core/ports/issues'
import type { Unsubscribe } from '_/core/ports/subscription'
import { Status } from '_/lib/async-status'
import { ok } from '_/lib/result'
import { anIssue } from '_/test/records'
import type { ActionEvent } from '_/types'
import { describe, expect, it } from 'vitest'
import { ContainerProvider, type Container } from './container'
import { useIssues } from './use-issues'

const connection: ConnectionPort = { onReconnect: () => () => {}, retry: () => {}, onStatusChange: () => () => {} }

const setup = (rows: Issue[]) => {
	let push: (issue: Issue, action: ActionEvent) => void = () => {}

	const issues = {
		list: async () => ok(rows as readonly Issue[]),
		subscribeToList: async (_project: string, update: (issue: Issue, action: ActionEvent) => void) => {
			push = update
			return ok((async () => {}) as Unsubscribe)
		},
	} as unknown as IssuesPort

	const container = { issues, connection } as unknown as Container
	const wrapper = ({ children }: { readonly children: React.ReactNode }) => (
		<ContainerProvider container={container}>{children}</ContainerProvider>
	)

	return { wrapper, push: (issue: Issue, action: ActionEvent) => act(() => push(issue, action)) }
}

const ids = (state: ReturnType<typeof useIssues>) => (state.status === Status.Ready ? state.issues.map(row => row.id) : [])

describe('useIssues', () => {
	it('orders priority by rank, not by the stored text', async () => {
		const io = setup([anIssue('m', { priority: 'medium' }), anIssue('h', { priority: 'high' }), anIssue('l', { priority: 'low' })])

		const { result } = renderHook(() => useIssues('p1', { sort: { field: 'priority', direction: 'desc' } }), {
			wrapper: io.wrapper,
		})

		await waitFor(() => expect(ids(result.current)).toEqual(['h', 'm', 'l']))
	})

	it('places a live insert where the sort puts it', async () => {
		const io = setup([anIssue('old', { createdAt: new Date(2026, 0, 1) })])

		const { result } = renderHook(() => useIssues('p1'), { wrapper: io.wrapper })
		await waitFor(() => expect(ids(result.current)).toEqual(['old']))

		await io.push(anIssue('new', { createdAt: new Date(2026, 0, 2) }), 'create')

		expect(ids(result.current)).toEqual(['new', 'old'])
	})

	it('moves a row when a live update changes its sort key', async () => {
		const io = setup([anIssue('a', { priority: 'high' }), anIssue('b', { priority: 'medium' })])

		const { result } = renderHook(() => useIssues('p1', { sort: { field: 'priority', direction: 'desc' } }), {
			wrapper: io.wrapper,
		})
		await waitFor(() => expect(ids(result.current)).toEqual(['a', 'b']))

		await io.push(anIssue('b', { priority: 'high', createdAt: new Date(2026, 0, 2) }), 'update')

		expect(ids(result.current)).toEqual(['b', 'a'])
	})
})

import { act, renderHook, waitFor } from '@testing-library/react'
import { projectId } from '_/core/domain/project'
import { shareId, type ShareLink, type ShareTarget } from '_/core/domain/share'
import type { ShareLinksPort } from '_/core/ports/share-links'
import { Status } from '_/lib/async-status'
import { err, ok } from '_/lib/result'
import { describe, expect, it, vi } from 'vitest'
import { ContainerProvider, type Container } from './container'
import { useShareLinks } from './use-share-links'

const target: ShareTarget = { kind: 'issue', id: 'i1', projectId: projectId('p1') }

const link = (id: string, label: string): ShareLink => ({
	id: shareId(id),
	label,
	token: 'k'.repeat(43),
	createdAt: new Date(0),
	lastAccessedAt: undefined,
})

const fakePort = (initial: ShareLink[]) => {
	let rows = initial
	const port = {
		list: vi.fn(async () => ok(rows as ReadonlyArray<ShareLink>)),
		create: vi.fn(async (_: ShareTarget, label: string) => {
			const created = link(`s${rows.length + 1}`, label)
			rows = [created, ...rows]
			return ok(created)
		}),
		revoke: vi.fn(async (id: ShareLink['id']) => {
			rows = rows.filter(row => row.id !== id)
			return ok(undefined)
		}),
	} satisfies ShareLinksPort

	return port
}

const render = (port: ShareLinksPort, skip = false) => {
	const container = { shareLinks: port } as unknown as Container
	const wrapper = ({ children }: { readonly children: React.ReactNode }) => (
		<ContainerProvider container={container}>{children}</ContainerProvider>
	)

	return renderHook(() => useShareLinks(target, skip), { wrapper })
}

describe('useShareLinks', () => {
	it('lists the target links', async () => {
		const port = fakePort([link('s1', 'vendor')])

		const { result } = render(port)

		await waitFor(() => expect(result.current.state).toEqual({ status: Status.Ready, data: [link('s1', 'vendor')] }))
		expect(port.list).toHaveBeenCalledWith(target)
	})

	it('shows a created link without reopening', async () => {
		const port = fakePort([])
		const { result } = render(port)
		await waitFor(() => expect(result.current.state.status).toBe(Status.Ready))

		await act(async () => {
			await result.current.create('vendor')
		})

		expect(port.create).toHaveBeenCalledWith(target, 'vendor')
		expect(result.current.state).toEqual({ status: Status.Ready, data: [link('s1', 'vendor')] })
	})

	it('drops a revoked link from the list', async () => {
		const port = fakePort([link('s1', 'vendor'), link('s2', 'reviewer')])
		const { result } = render(port)
		await waitFor(() => expect(result.current.state.status).toBe(Status.Ready))

		await act(async () => {
			await result.current.revoke(shareId('s1'))
		})

		expect(result.current.state).toEqual({ status: Status.Ready, data: [link('s2', 'reviewer')] })
	})

	it('hands back a failed create and leaves the list alone', async () => {
		const port = fakePort([link('s1', 'vendor')])
		port.create.mockResolvedValueOnce(err(new Error('Could not create the link: forbidden')))
		const { result } = render(port)
		await waitFor(() => expect(result.current.state.status).toBe(Status.Ready))

		let outcome: Awaited<ReturnType<typeof result.current.create>> | undefined
		await act(async () => {
			outcome = await result.current.create('vendor')
		})

		expect(outcome).toMatchObject({ success: false, error: { message: 'Could not create the link: forbidden' } })
		expect(port.list).toHaveBeenCalledTimes(1)
	})

	it('loads nothing while skipped', async () => {
		const port = fakePort([link('s1', 'vendor')])

		render(port, true)
		await act(async () => {})

		expect(port.list).not.toHaveBeenCalled()
	})
})

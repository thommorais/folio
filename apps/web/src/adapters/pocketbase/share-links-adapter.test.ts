import { afterEach, describe, expect, it, vi } from 'vitest'
import { projectId } from '_/core/domain/project'
import { shareId, type ShareTarget } from '_/core/domain/share'

const getFullList = vi.fn()
const create = vi.fn()
const remove = vi.fn()
const authStore: { record: { id: string } | null } = { record: { id: 'u-me' } }

vi.mock('./client', () => ({
	getPocketBaseClient: () => ({
		authStore,
		filter: (expr: string, params: Record<string, string>) => expr.replace('{:id}', `'${params.id}'`),
		collection: () => ({ getFullList, create, delete: remove }),
	}),
}))

const { createShareLinksAdapter } = await import('./share-links-adapter')

const ticket: ShareTarget = { kind: 'issue', id: 'i1', projectId: projectId('p1') }
const plan: ShareTarget = { kind: 'plan', id: 'pl1', projectId: projectId('p1') }

const record = (id: string, label: string, lastAccessed = '') => ({
	id,
	label,
	token: 'k'.repeat(43),
	created: '2026-09-01 10:00:00.000Z',
	last_accessed_at: lastAccessed,
})

describe('share links adapter', () => {
	afterEach(() => {
		vi.resetAllMocks()
	})

	it('lists the links of a ticket, newest first', async () => {
		getFullList.mockResolvedValueOnce([record('s2', 'vendor', '2026-09-02 08:00:00.000Z'), record('s1', 'reviewer')])

		const result = await createShareLinksAdapter().list(ticket)

		expect(getFullList).toHaveBeenCalledWith(expect.objectContaining({ filter: "issue = 'i1'", sort: '-created' }))
		expect(result).toEqual({
			success: true,
			value: [
				{
					id: shareId('s2'),
					label: 'vendor',
					token: 'k'.repeat(43),
					createdAt: new Date('2026-09-01T10:00:00Z'),
					lastAccessedAt: new Date('2026-09-02T08:00:00Z'),
				},
				{
					id: shareId('s1'),
					label: 'reviewer',
					token: 'k'.repeat(43),
					createdAt: new Date('2026-09-01T10:00:00Z'),
					lastAccessedAt: undefined,
				},
			],
		})
	})

	it('lists the links of a plan by its own field', async () => {
		getFullList.mockResolvedValueOnce([])

		await createShareLinksAdapter().list(plan)

		expect(getFullList).toHaveBeenCalledWith(expect.objectContaining({ filter: "plan = 'pl1'" }))
	})

	it('files a new link under the caller and leaves the token to the server', async () => {
		create.mockResolvedValueOnce(record('s3', 'vendor'))

		const result = await createShareLinksAdapter().create(ticket, '  vendor  ')

		expect(create).toHaveBeenCalledWith({ project: 'p1', issue: 'i1', label: 'vendor', created_by: 'u-me' })
		expect(result).toMatchObject({ success: true, value: { id: shareId('s3'), label: 'vendor' } })
	})

	it('points a plan link at the plan', async () => {
		create.mockResolvedValueOnce(record('s4', 'reviewer'))

		await createShareLinksAdapter().create(plan, 'reviewer')

		expect(create).toHaveBeenCalledWith({ project: 'p1', plan: 'pl1', label: 'reviewer', created_by: 'u-me' })
	})

	it('revokes a link by deleting it', async () => {
		remove.mockResolvedValueOnce(true)

		const result = await createShareLinksAdapter().revoke(shareId('s1'))

		expect(remove).toHaveBeenCalledWith('s1')
		expect(result).toEqual({ success: true, value: undefined })
	})

	it.each([
		['list', () => createShareLinksAdapter().list(ticket), getFullList, 'Failed to load share links'],
		['create', () => createShareLinksAdapter().create(ticket, 'vendor'), create, 'Could not create the link'],
		['revoke', () => createShareLinksAdapter().revoke(shareId('s1')), remove, 'Could not revoke the link'],
	])('reports a failed %s', async (_, call, method, message) => {
		method.mockRejectedValueOnce(new Error('Failed to create record.'))

		const result = await call()

		expect(result).toMatchObject({ success: false, error: { message: expect.stringContaining(message) } })
	})
})

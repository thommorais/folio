import { afterEach, describe, expect, it, vi } from 'vitest'
import { createSharesAdapter } from './shares-adapter'

const issue = (id: string, title: string) => ({
	id,
	kind: 'todo',
	title,
	body: '',
	status: 'open',
	priority: 'medium',
	tags: [],
	updated_at: '2026-09-01T10:00:00Z',
})

const issueShare = {
	kind: 'issue',
	label: 'vendor',
	brief: {
		issue: { ...issue('t1', 'Checkout flow'), kind: 'ticket', body: 'use stripe', external_ref: 'WEL-12' },
		children: [issue('c1', 'wire webhooks')],
		plans: [
			{
				id: 'p1',
				title: 'Payments',
				status: 'active',
				tags: [],
				progress: { total: 0, done: 0, percent: 0 },
				updated_at: '2026-09-01T10:00:00Z',
			},
		],
		journal: [{ id: 'j1', kind: 'journal', title: 'Picked stripe', body: 'cheaper', created_at: '2026-09-02T10:00:00Z' }],
		docs: [],
		cycles: [{ id: 'cy1', ordinal: 1, phase: 'check', closed_at: '2026-09-03T10:00:00Z' }],
	},
}

const planShare = {
	kind: 'plan',
	label: 'reviewer',
	plan: {
		id: 'p1',
		title: 'Migrate auth',
		goal: 'drop sessions',
		status: 'draft',
		tags: ['auth'],
		progress: { total: 1, done: 0, percent: 0 },
		updated_at: '2026-09-01T10:00:00Z',
	},
	todos: [issue('c1', 'add tokens')],
}

const respond = (status: number, body: unknown) =>
	vi.stubGlobal(
		'fetch',
		vi.fn(async () => new Response(typeof body === 'string' ? body : JSON.stringify(body), { status })),
	)

describe('shares adapter', () => {
	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it('asks for the token without credentials', async () => {
		respond(200, planShare)

		await createSharesAdapter().open('a/b')

		expect(fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/share\/a%2Fb$/), { credentials: 'omit' })
	})

	it('maps a ticket share', async () => {
		respond(200, issueShare)

		const result = await createSharesAdapter().open('tok')

		expect(result).toMatchObject({
			success: true,
			value: {
				kind: 'issue',
				label: 'vendor',
				issue: { title: 'Checkout flow', body: 'use stripe', externalRef: 'WEL-12', updatedAt: new Date('2026-09-01T10:00:00Z') },
				todos: [{ id: 'c1', title: 'wire webhooks' }],
				plans: [{ id: 'p1', title: 'Payments', goal: '' }],
				journal: [{ title: 'Picked stripe', createdAt: new Date('2026-09-02T10:00:00Z') }],
				docs: [],
				cycles: [{ ordinal: 1, resolution: '', closedAt: new Date('2026-09-03T10:00:00Z') }],
			},
		})
	})

	it('maps a plan share', async () => {
		respond(200, planShare)

		const result = await createSharesAdapter().open('tok')

		expect(result).toMatchObject({
			success: true,
			value: {
				kind: 'plan',
				label: 'reviewer',
				plan: { title: 'Migrate auth', goal: 'drop sessions', progress: { total: 1, done: 0 } },
				todos: [{ id: 'c1', title: 'add tokens' }],
			},
		})
	})

	it('treats a 404 as a link that opens nothing', async () => {
		respond(404, { message: 'not found' })

		expect(await createSharesAdapter().open('tok')).toEqual({ success: true, value: undefined })
	})

	it.each([
		['a server error', 500, { message: 'boom' }],
		['a body that is not JSON', 200, '<html>'],
		['an unexpected shape', 200, { kind: 'issue', label: 'vendor' }],
	])('fails on %s', async (_, status, body) => {
		respond(status, body)

		expect((await createSharesAdapter().open('tok')).success).toBe(false)
	})

	it('fails when the server cannot be reached', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(async () => {
				throw new TypeError('Failed to fetch')
			}),
		)

		expect(await createSharesAdapter().open('tok')).toMatchObject({
			success: false,
			error: { message: 'Cannot reach the server.' },
		})
	})
})

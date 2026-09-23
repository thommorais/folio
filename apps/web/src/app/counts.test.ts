import { err, ok } from '_/lib/result'
import { describe, expect, it } from 'vitest'
import { collectCounts } from './counts'
import { Status } from '_/lib/async-status'

describe('collectCounts', () => {
	it('pairs each total with the entity it was counted for', () => {
		const state = collectCounts([ok(2), ok(3), ok(13), ok(6)])

		expect(state).toEqual({
			status: Status.Ready,
			counts: { tickets: 2, plans: 3, todos: 13, journal: 6 },
		})
	})

	it('fails the whole row when one count fails, so no tile shows a wrong number', () => {
		const state = collectCounts([ok(2), ok(3), err(new Error('Failed to count todos: offline')), ok(6)])

		expect(state).toEqual({ status: Status.Failed, message: 'Failed to count todos: offline' })
	})

	it('reports the first failure when several fail', () => {
		const state = collectCounts([ok(2), err(new Error('plans blew up')), err(new Error('todos blew up')), ok(6)])

		expect(state).toEqual({ status: Status.Failed, message: 'plans blew up' })
	})

	it('keeps a zero count as a real total rather than treating it as missing', () => {
		const state = collectCounts([ok(0), ok(0), ok(0), ok(0)])

		expect(state).toEqual({
			status: Status.Ready,
			counts: { tickets: 0, plans: 0, todos: 0, journal: 0 },
		})
	})
})

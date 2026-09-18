import { describe, expect, it } from 'vitest'
import { cycleProgress, newestFirst } from './cycle-progress'
import type { Cycle, CycleId } from './cycle'

const cycle = (ordinal: number, over: Partial<Cycle> = {}): Cycle =>
	({
		id: `c${ordinal}` as CycleId,
		projectId: 'p1',
		ticketId: 't1',
		ordinal,
		phase: 'plan',
		resolution: '',
		createdBy: undefined,
		createdAt: new Date('2026-01-01'),
		updatedAt: new Date('2026-01-01'),
		closedAt: undefined,
		...over,
	}) as Cycle

describe('cycleProgress', () => {
	it('marks the phases up to the current one as reached', () => {
		const got = cycleProgress(cycle(1, { phase: 'check' }))

		expect(got.map(step => [step.phase, step.reached])).toEqual([
			['plan', true],
			['do', true],
			['check', true],
			['act', false],
		])
	})

	it('marks only the current phase as current', () => {
		const got = cycleProgress(cycle(1, { phase: 'do' }))

		expect(got.filter(step => step.current).map(step => step.phase)).toEqual(['do'])
	})

	it('reaches only plan on a fresh cycle', () => {
		const got = cycleProgress(cycle(1, { phase: 'plan' }))

		expect(got.filter(step => step.reached).map(step => step.phase)).toEqual(['plan'])
	})

	it('reaches every phase at act', () => {
		const got = cycleProgress(cycle(1, { phase: 'act' }))

		expect(got.every(step => step.reached)).toBe(true)
	})

	// A resolved cycle is finished whatever phase it stopped at, so nothing in
	// it is still in progress.
	it('leaves no phase current once the cycle is resolved', () => {
		const got = cycleProgress(cycle(1, { phase: 'do', closedAt: new Date('2026-02-01') }))

		expect(got.some(step => step.current)).toBe(false)
		expect(got.filter(step => step.reached).map(step => step.phase)).toEqual(['plan', 'do'])
	})
})

describe('newestFirst', () => {
	it('puts the highest ordinal first', () => {
		const got = newestFirst([cycle(1), cycle(3), cycle(2)])

		expect(got.map(one => one.ordinal)).toEqual([3, 2, 1])
	})

	it('leaves the input untouched', () => {
		const cycles = [cycle(1), cycle(2)]
		newestFirst(cycles)

		expect(cycles.map(one => one.ordinal)).toEqual([1, 2])
	})
})

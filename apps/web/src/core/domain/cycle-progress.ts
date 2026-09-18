import type { Cycle, Phase } from './cycle'
import { PHASES, isResolved } from './cycle'

export type PhaseStep = {
	readonly phase: Phase
	// Reached covers the phase the cycle stopped at and everything before it.
	readonly reached: boolean
	// Where the work sits now. A resolved cycle is nowhere: it is finished at
	// whatever phase it stopped, which is not the same as still being there.
	readonly current: boolean
}

export const cycleProgress = (cycle: Cycle): readonly PhaseStep[] => {
	const at = PHASES.indexOf(cycle.phase)
	const resolved = isResolved(cycle)

	return PHASES.map((phase, index) => ({
		phase,
		reached: index <= at,
		current: !resolved && index === at,
	}))
}

// Rounds read newest first: the open one is what a reader wants, and the
// history below it is context.
export const newestFirst = (cycles: readonly Cycle[]): readonly Cycle[] =>
	[...cycles].sort((a, b) => b.ordinal - a.ordinal)

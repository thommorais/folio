import type { Result } from '_/lib/result'
import { Status } from '_/lib/async-status'

export const entities = ['tickets', 'plans', 'todos', 'journal'] as const

export type Entity = (typeof entities)[number]

export type Counts = Readonly<Record<Entity, number>>

export type CountsState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly counts: Counts }
	| { readonly status: typeof Status.Failed; readonly message: string }

// A partial row would show a stale tile beside a fresh one with nothing to
// tell them apart, so one failed count fails the whole row.
export const collectCounts = (results: readonly Result<number>[]): CountsState => {
	const totals: number[] = []

	for (const result of results) {
		if (!result.success) {
			return { status: Status.Failed, message: result.error.message }
		}
		totals.push(result.value)
	}

	const [tickets = 0, plans = 0, todos = 0, journal = 0] = totals

	return { status: Status.Ready, counts: { tickets, plans, todos, journal } }
}

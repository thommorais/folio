import type { Result } from '_/lib/result'

export const entities = ['tickets', 'plans', 'todos', 'journal'] as const

export type Entity = (typeof entities)[number]

export type Counts = Readonly<Record<Entity, number>>

export type CountsState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly counts: Counts }
	| { readonly status: 'failed'; readonly message: string }

// A partial row would show a stale tile beside a fresh one with nothing to
// tell them apart, so one failed count fails the whole row.
export const collectCounts = (results: readonly Result<number>[]): CountsState => {
	const totals: number[] = []

	for (const result of results) {
		if (!result.success) {
			return { status: 'failed', message: result.error.message }
		}
		totals.push(result.value)
	}

	const [tickets = 0, plans = 0, todos = 0, journal = 0] = totals

	return { status: 'ready', counts: { tickets, plans, todos, journal } }
}

import type { Issue } from './issue'
import type { Plan } from './plan'

export type GroupTarget =
	| { readonly kind: 'plan'; readonly id: string; readonly plan: Plan }
	| { readonly kind: 'ticket'; readonly id: string; readonly ticket: Issue }

export type Group<T> = {
	readonly target: GroupTarget | undefined
	readonly items: readonly T[]
}

const group = <T>(items: readonly T[], targetOf: (item: T) => GroupTarget | undefined): readonly Group<T>[] => {
	const grouped = new Map<string, { target: GroupTarget; items: T[] }>()
	const loose: T[] = []

	for (const item of items) {
		const target = targetOf(item)

		if (target === undefined) {
			loose.push(item)
			continue
		}

		const key = `${target.kind}:${target.id}`
		const bucket = grouped.get(key)

		if (bucket === undefined) {
			grouped.set(key, { target, items: [item] })
		} else {
			bucket.items.push(item)
		}
	}

	const groups: Group<T>[] = [...grouped.values()]

	return loose.length > 0 ? [...groups, { target: undefined, items: loose }] : groups
}

const ticketTarget = (tickets: ReadonlyMap<string, Issue>, id: string | undefined): GroupTarget | undefined => {
	const found = id === undefined ? undefined : tickets.get(id)

	return found === undefined ? undefined : { kind: 'ticket', id: found.id, ticket: found }
}

export const groupTodos = (
	todos: readonly Issue[],
	plans: readonly Plan[],
	tickets: readonly Issue[],
): readonly Group<Issue>[] => {
	const planById = new Map<string, Plan>(plans.map(plan => [plan.id, plan]))
	const ticketById = new Map<string, Issue>(tickets.map(ticket => [ticket.id, ticket]))

	return group(todos, todo => {
		const found = todo.planId === undefined ? undefined : planById.get(todo.planId)

		return found === undefined ? ticketTarget(ticketById, todo.parentId) : { kind: 'plan', id: found.id, plan: found }
	})
}

export const groupPlans = (plans: readonly Plan[], tickets: readonly Issue[]): readonly Group<Plan>[] => {
	const ticketById = new Map<string, Issue>(tickets.map(ticket => [ticket.id, ticket]))

	return group(plans, plan => ticketTarget(ticketById, plan.ticketId))
}

import type { Issue } from './issue'
import { isTerminal } from './issue'

export type Partitioned = {
	readonly frontier: readonly Issue[]
	readonly blocked: readonly Issue[]
	readonly claimed: readonly Issue[]
	readonly done: readonly Issue[]
}

const oldestFirst = (a: Issue, b: Issue): number => a.createdAt.getTime() - b.createdAt.getTime()

export const partitionChildren = (children: readonly Issue[]): Partitioned => {
	const frontier: Issue[] = []
	const blocked: Issue[] = []
	const claimed: Issue[] = []
	const done: Issue[] = []

	for (const child of children) {
		if (isTerminal(child.status)) {
			done.push(child)
		} else if (child.blocked) {
			blocked.push(child)
		} else if (child.assignee !== undefined) {
			claimed.push(child)
		} else {
			frontier.push(child)
		}
	}

	return {
		frontier: frontier.sort(oldestFirst),
		blocked: blocked.sort(oldestFirst),
		claimed: claimed.sort(oldestFirst),
		done: done.sort(oldestFirst),
	}
}

import type { Edge } from './graph'
import { LINK_KIND, type IssueId } from './issue'

// Everything reachable from a node along its edges, in either direction: what
// it waits on, what waits on it, where it sits and what it is related to.
// Hovering one node dims the rest, and this is the set that stays lit.
export const connectedTo = (edges: readonly Edge[], focus: IssueId): ReadonlySet<IssueId> => {
	const neighbours = new Map<IssueId, IssueId[]>()

	const join = (from: IssueId, to: IssueId) => {
		const bucket = neighbours.get(from)

		if (bucket === undefined) {
			neighbours.set(from, [to])
		} else {
			bucket.push(to)
		}
	}

	for (const edge of edges) {
		if (edge.kind === LINK_KIND.PARENT) {
			// Upwards only. A parent places a node, but walking back down it
			// would light up every sibling, which is the bulk of a map and
			// exactly what the highlight is meant to dim.
			join(edge.to, edge.from)
			continue
		}

		join(edge.from, edge.to)
		join(edge.to, edge.from)
	}

	const reached = new Set<IssueId>([focus])
	const queue: IssueId[] = [focus]

	while (queue.length > 0) {
		const current = queue.shift()

		if (current === undefined) continue

		for (const next of neighbours.get(current) ?? []) {
			if (reached.has(next)) continue

			reached.add(next)
			queue.push(next)
		}
	}

	return reached
}

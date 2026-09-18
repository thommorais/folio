import type { Issue, IssueId, LinkKind } from './issue'

export type Edge = {
	readonly from: IssueId
	readonly to: IssueId
	readonly kind: LinkKind
}

export type PlacedNode = {
	readonly issue: Issue
	// Distance from the top of the graph: the longest path along parent and
	// blocker links, so an edge always points downwards.
	readonly rank: number
	// Position within the rank, left to right.
	readonly order: number
}

export type Graph = {
	readonly nodes: readonly PlacedNode[]
	readonly edges: readonly Edge[]
}

export const subtreeOf = (issues: readonly Issue[], rootId: IssueId): readonly Issue[] => {
	const root = issues.find(issue => issue.id === rootId)

	if (root === undefined) return []

	const childrenOf = new Map<IssueId, Issue[]>()

	for (const issue of issues) {
		if (issue.parentId === undefined) continue

		const bucket = childrenOf.get(issue.parentId)

		if (bucket === undefined) {
			childrenOf.set(issue.parentId, [issue])
		} else {
			bucket.push(issue)
		}
	}

	const collected: Issue[] = []
	// A parent cycle would otherwise walk forever.
	const seen = new Set<IssueId>()
	const queue: Issue[] = [root]

	while (queue.length > 0) {
		const issue = queue.shift()

		if (issue === undefined || seen.has(issue.id)) continue

		seen.add(issue.id)
		collected.push(issue)
		queue.push(...(childrenOf.get(issue.id) ?? []))
	}

	return collected
}

const edgesOf = (issues: readonly Issue[]): readonly Edge[] => {
	const present = new Set(issues.map(issue => issue.id))
	const edges: Edge[] = []
	// A relation is stored on both ends; one edge is enough to draw it.
	const drawn = new Set<string>()

	for (const issue of issues) {
		if (issue.parentId !== undefined && present.has(issue.parentId)) {
			edges.push({ from: issue.parentId, to: issue.id, kind: 'parent' })
		}

		for (const blocker of issue.dependsOn) {
			if (present.has(blocker)) edges.push({ from: blocker, to: issue.id, kind: 'blocks' })
		}

		for (const other of issue.relatedTo) {
			if (!present.has(other)) continue

			const key = [issue.id, other].sort().join('|')

			if (drawn.has(key)) continue

			drawn.add(key)
			edges.push({ from: issue.id, to: other, kind: 'relates' })
		}
	}

	return edges
}

// Longest-path ranking over the downward edges. A cycle among them is refused
// at write time, but a stale or hand-edited set can still carry one, so the
// walk tracks its own stack and drops the edge that closes the loop rather
// than recursing forever.
const ranksOf = (issues: readonly Issue[], edges: readonly Edge[]): ReadonlyMap<IssueId, number> => {
	const incoming = new Map<IssueId, IssueId[]>()

	for (const edge of edges) {
		if (edge.kind === 'relates') continue

		const bucket = incoming.get(edge.to)

		if (bucket === undefined) {
			incoming.set(edge.to, [edge.from])
		} else {
			bucket.push(edge.from)
		}
	}

	const ranks = new Map<IssueId, number>()
	const onStack = new Set<IssueId>()

	const rankOf = (id: IssueId): number => {
		const known = ranks.get(id)

		if (known !== undefined) return known
		if (onStack.has(id)) return 0

		onStack.add(id)

		let rank = 0

		for (const parent of incoming.get(id) ?? []) {
			rank = Math.max(rank, rankOf(parent) + 1)
		}

		onStack.delete(id)
		ranks.set(id, rank)

		return rank
	}

	for (const issue of issues) rankOf(issue.id)

	return ranks
}

export const layout = (issues: readonly Issue[]): Graph => {
	const edges = edgesOf(issues)
	const ranks = ranksOf(issues, edges)

	const byRank = new Map<number, Issue[]>()

	for (const issue of issues) {
		const rank = ranks.get(issue.id) ?? 0
		const bucket = byRank.get(rank)

		if (bucket === undefined) {
			byRank.set(rank, [issue])
		} else {
			bucket.push(issue)
		}
	}

	const position = new Map(issues.map((issue, index) => [issue.id, index]))

	// Ordering by the mean position of a node's parents keeps an edge as close
	// to vertical as it can be, which is what stops the picture looking woven.
	const barycentre = (issue: Issue): number => {
		const above = edges.filter(edge => edge.to === issue.id && edge.kind !== 'relates')

		if (above.length === 0) return position.get(issue.id) ?? 0

		const sum = above.reduce((total, edge) => total + (position.get(edge.from) ?? 0), 0)

		return sum / above.length
	}

	const nodes: PlacedNode[] = []

	for (const [rank, bucket] of [...byRank.entries()].sort(([a], [b]) => a - b)) {
		const ordered = [...bucket].sort((a, b) => barycentre(a) - barycentre(b))

		for (const [order, issue] of ordered.entries()) {
			nodes.push({ issue, rank, order })
		}
	}

	return { nodes, edges }
}

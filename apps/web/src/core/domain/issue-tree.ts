import type { Issue, IssueId } from './issue'

export type IssueRow = {
	readonly issue: Issue
	readonly depth: number
	// The last child of its parent, so the connector can close the run.
	readonly isLast: boolean
	// An ancestor pulled in to place a match, not a match itself.
	readonly isContext: boolean
	// One flag per ancestor level: does that ancestor have siblings after it?
	// Drives whether a vertical guide continues past this row.
	readonly guides: readonly boolean[]
}

export type TreeOptions = {
	// Tickets that may be pulled in as ancestors of a match. A filtered list
	// only carries matches, so the parents needed to place them come from here.
	readonly context?: readonly Issue[]
}

type Node = {
	readonly issue: Issue
	readonly children: Node[]
	readonly isContext: boolean
}

const rootKey = Symbol('root')

type Key = IssueId | typeof rootKey

// Walks up from a match through the context pool, collecting every ancestor
// needed to place it. Stops at a issue already in the tree or a broken link.
const ancestorsOf = (
	issue: Issue,
	context: ReadonlyMap<IssueId, Issue>,
	included: ReadonlySet<IssueId>,
): readonly Issue[] => {
	const chain: Issue[] = []
	const seen = new Set<IssueId>([issue.id])

	let parentId = issue.parentId

	while (parentId !== undefined && !included.has(parentId) && !seen.has(parentId)) {
		seen.add(parentId)

		const parent = context.get(parentId)

		if (parent === undefined) break

		chain.push(parent)
		parentId = parent.parentId
	}

	return chain.reverse()
}

export const buildIssueTree = (tickets: readonly Issue[], options: TreeOptions = {}): readonly IssueRow[] => {
	const matched = new Set(tickets.map(issue => issue.id))
	const contextById = new Map((options.context ?? []).map(issue => [issue.id, issue]))

	// Ancestors come first so a parent is always positioned before its child.
	const ordered: { readonly issue: Issue; readonly isContext: boolean }[] = []
	const placed = new Set<IssueId>()

	for (const issue of tickets) {
		for (const ancestor of ancestorsOf(issue, contextById, placed)) {
			if (placed.has(ancestor.id)) continue

			placed.add(ancestor.id)
			ordered.push({ issue: ancestor, isContext: !matched.has(ancestor.id) })
		}

		if (placed.has(issue.id)) continue

		placed.add(issue.id)
		ordered.push({ issue, isContext: false })
	}

	const nodes = new Map<IssueId, Node>(
		ordered.map(entry => [entry.issue.id, { issue: entry.issue, children: [], isContext: entry.isContext }]),
	)

	const childrenOf = new Map<Key, Node[]>([[rootKey, []]])

	for (const entry of ordered) {
		const node = nodes.get(entry.issue.id)

		if (node === undefined) continue

		// A parent outside the tree (filtered away, or a dangling id) would strand
		// the row, so it is attached at the root instead of dropped.
		const parentId = entry.issue.parentId
		const key: Key = parentId !== undefined && nodes.has(parentId) ? parentId : rootKey

		const bucket = childrenOf.get(key)

		if (bucket === undefined) {
			childrenOf.set(key, [node])
		} else {
			bucket.push(node)
		}
	}

	const rows: IssueRow[] = []
	// A parent cycle would otherwise recurse forever.
	const visited = new Set<IssueId>()

	const walk = (siblings: readonly Node[], depth: number, guides: readonly boolean[]) => {
		for (const [index, node] of siblings.entries()) {
			if (visited.has(node.issue.id)) continue

			visited.add(node.issue.id)

			const isLast = index === siblings.length - 1

			rows.push({
				issue: node.issue,
				depth,
				isLast,
				isContext: node.isContext,
				guides,
			})

			const children = childrenOf.get(node.issue.id)

			if (children !== undefined && children.length > 0) {
				walk(children, depth + 1, [...guides, !isLast])
			}
		}
	}

	walk(childrenOf.get(rootKey) ?? [], 0, [])

	// A cycle leaves its members unreachable from the root; list them flat so
	// nothing silently disappears.
	for (const entry of ordered) {
		if (visited.has(entry.issue.id)) continue

		visited.add(entry.issue.id)
		rows.push({ issue: entry.issue, depth: 0, isLast: true, isContext: entry.isContext, guides: [] })
	}

	return rows
}

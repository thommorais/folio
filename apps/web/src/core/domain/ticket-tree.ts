import type { Ticket, TicketId } from './ticket'

export type TicketRow = {
	readonly ticket: Ticket
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
	readonly context?: readonly Ticket[]
}

type Node = {
	readonly ticket: Ticket
	readonly children: Node[]
	readonly isContext: boolean
}

const rootKey = Symbol('root')

type Key = TicketId | typeof rootKey

// Walks up from a match through the context pool, collecting every ancestor
// needed to place it. Stops at a ticket already in the tree or a broken link.
const ancestorsOf = (
	ticket: Ticket,
	context: ReadonlyMap<TicketId, Ticket>,
	included: ReadonlySet<TicketId>,
): readonly Ticket[] => {
	const chain: Ticket[] = []
	const seen = new Set<TicketId>([ticket.id])

	let parentId = ticket.parentId

	while (parentId !== undefined && !included.has(parentId) && !seen.has(parentId)) {
		seen.add(parentId)

		const parent = context.get(parentId)

		if (parent === undefined) break

		chain.push(parent)
		parentId = parent.parentId
	}

	return chain.reverse()
}

export const buildTicketTree = (tickets: readonly Ticket[], options: TreeOptions = {}): readonly TicketRow[] => {
	const matched = new Set(tickets.map(ticket => ticket.id))
	const contextById = new Map((options.context ?? []).map(ticket => [ticket.id, ticket]))

	// Ancestors come first so a parent is always positioned before its child.
	const ordered: { readonly ticket: Ticket; readonly isContext: boolean }[] = []
	const placed = new Set<TicketId>()

	for (const ticket of tickets) {
		for (const ancestor of ancestorsOf(ticket, contextById, placed)) {
			if (placed.has(ancestor.id)) continue

			placed.add(ancestor.id)
			ordered.push({ ticket: ancestor, isContext: !matched.has(ancestor.id) })
		}

		if (placed.has(ticket.id)) continue

		placed.add(ticket.id)
		ordered.push({ ticket, isContext: false })
	}

	const nodes = new Map<TicketId, Node>(
		ordered.map(entry => [entry.ticket.id, { ticket: entry.ticket, children: [], isContext: entry.isContext }]),
	)

	const childrenOf = new Map<Key, Node[]>([[rootKey, []]])

	for (const entry of ordered) {
		const node = nodes.get(entry.ticket.id)

		if (node === undefined) continue

		// A parent outside the tree (filtered away, or a dangling id) would strand
		// the row, so it is attached at the root instead of dropped.
		const parentId = entry.ticket.parentId
		const key: Key = parentId !== undefined && nodes.has(parentId) ? parentId : rootKey

		const bucket = childrenOf.get(key)

		if (bucket === undefined) {
			childrenOf.set(key, [node])
		} else {
			bucket.push(node)
		}
	}

	const rows: TicketRow[] = []
	// A parent cycle would otherwise recurse forever.
	const visited = new Set<TicketId>()

	const walk = (siblings: readonly Node[], depth: number, guides: readonly boolean[]) => {
		for (const [index, node] of siblings.entries()) {
			if (visited.has(node.ticket.id)) continue

			visited.add(node.ticket.id)

			const isLast = index === siblings.length - 1

			rows.push({
				ticket: node.ticket,
				depth,
				isLast,
				isContext: node.isContext,
				guides,
			})

			const children = childrenOf.get(node.ticket.id)

			if (children !== undefined && children.length > 0) {
				walk(children, depth + 1, [...guides, !isLast])
			}
		}
	}

	walk(childrenOf.get(rootKey) ?? [], 0, [])

	// A cycle leaves its members unreachable from the root; list them flat so
	// nothing silently disappears.
	for (const entry of ordered) {
		if (visited.has(entry.ticket.id)) continue

		visited.add(entry.ticket.id)
		rows.push({ ticket: entry.ticket, depth: 0, isLast: true, isContext: entry.isContext, guides: [] })
	}

	return rows
}

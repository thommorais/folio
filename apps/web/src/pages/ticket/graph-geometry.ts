import type { Edge, Graph } from '_/core/domain/graph'
import type { Issue, IssueId, LinkKind } from '_/core/domain/issue'

export const NODE = {
	width: 200,
	height: 56,
	gapX: 32,
	gapY: 52,
} as const

export type Box = {
	readonly id: IssueId
	readonly issue: Issue
	readonly x: number
	readonly y: number
}

export type Point = { readonly x: number; readonly y: number }

export type Line = {
	readonly id: string
	readonly kind: LinkKind
	readonly from: Point
	readonly to: Point
	readonly between: readonly [IssueId, IssueId]
}

export type Placed = {
	readonly boxes: readonly Box[]
	readonly lines: readonly Line[]
	readonly width: number
	readonly height: number
}

const rowWidth = (count: number): number => count * NODE.width + (count - 1) * NODE.gapX

// An edge leaves the bottom edge of its source and arrives at the top of its
// target, so the arrowhead always meets the box square rather than at a corner.
const exit = (box: Box): Point => ({ x: box.x + NODE.width / 2, y: box.y + NODE.height })

const entry = (box: Box): Point => ({ x: box.x + NODE.width / 2, y: box.y })

// A relation joins two nodes that often share a rank, where bottom-to-top would
// double back on itself. It runs side to side instead.
const sideways = (from: Box, to: Box): readonly [Point, Point] => {
	const [left, right] = from.x <= to.x ? [from, to] : [to, from]

	return [
		{ x: left.x + NODE.width, y: left.y + NODE.height / 2 },
		{ x: right.x, y: right.y + NODE.height / 2 },
	]
}

const endpoints = (edge: Edge, from: Box, to: Box): readonly [Point, Point] => {
	if (edge.kind === 'relates') return sideways(from, to)

	return from.y <= to.y ? [exit(from), entry(to)] : [entry(from), exit(to)]
}

export const place = (graph: Graph): Placed => {
	if (graph.nodes.length === 0) return { boxes: [], lines: [], width: 0, height: 0 }

	const ranks = new Map<number, typeof graph.nodes>()

	for (const node of graph.nodes) {
		const bucket = ranks.get(node.rank)
		ranks.set(node.rank, bucket === undefined ? [node] : [...bucket, node])
	}

	const widest = Math.max(...[...ranks.values()].map(bucket => rowWidth(bucket.length)))

	const boxes: Box[] = []

	for (const [rank, bucket] of ranks) {
		// Each rank is centred against the widest one, so a single node above a
		// row of four sits over the middle of them.
		const offset = (widest - rowWidth(bucket.length)) / 2

		for (const node of bucket) {
			boxes.push({
				id: node.issue.id,
				issue: node.issue,
				x: offset + node.order * (NODE.width + NODE.gapX),
				y: rank * (NODE.height + NODE.gapY),
			})
		}
	}

	const byId = new Map(boxes.map(box => [box.id, box]))

	const lines: Line[] = []

	for (const edge of graph.edges) {
		const from = byId.get(edge.from)
		const to = byId.get(edge.to)

		if (from === undefined || to === undefined) continue

		const [start, end] = endpoints(edge, from, to)

		lines.push({
			id: `${edge.kind}:${edge.from}:${edge.to}`,
			kind: edge.kind,
			from: start,
			to: end,
			between: [edge.from, edge.to],
		})
	}

	// Measured from the boxes rather than the widest rank: a node whose order
	// leaves a gap on its own rank still has to fit on the canvas.
	const right = Math.max(...boxes.map(box => box.x)) + NODE.width
	const depth = Math.max(...boxes.map(box => box.y)) + NODE.height

	return { boxes, lines, width: right, height: depth }
}

import { useNavigate } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { useState } from 'react'
import { layout, subtreeOf } from '_/core/domain/graph'
import { connectedTo } from '_/core/domain/graph-path'
import { LINK_KIND, type Issue, type IssueId } from '_/core/domain/issue'
import { useScope } from '_/routing/use-scope'
import { NODE, place, type Line } from './graph-geometry'
import { GraphNode } from './graph-node'

type Props = {
	readonly project: string
	readonly mapId: IssueId
	readonly issues: readonly Issue[]
}

// An edge bends once, halfway down the gap between the ranks, so a run of them
// reads as a set of tracks rather than a fan of diagonals.
const pathOf = (line: Line): string => {
	if (line.kind === LINK_KIND.RELATES) {
		const lift = Math.max(24, Math.abs(line.to.x - line.from.x) / 3)
		return `M ${line.from.x} ${line.from.y} C ${line.from.x + lift} ${line.from.y}, ${line.to.x - lift} ${line.to.y}, ${line.to.x} ${line.to.y}`
	}

	// Turn in the gap directly below the source, then travel sideways there
	// before dropping. An edge spanning more than one rank would otherwise
	// descend through whatever sits between the two.
	const turn = line.from.y + NODE.gapY / 2

	return `M ${line.from.x} ${line.from.y} V ${turn} H ${line.to.x} V ${line.to.y}`
}

const WayfinderGraph = ({ project, mapId, issues }: Props) => {
	const navigate = useNavigate()
	const { client, domain } = useScope()
	const [focus, setFocus] = useState<IssueId | undefined>(undefined)

	const full = layout(subtreeOf(issues, mapId))

	// A node reached by a blocker is already placed by that edge. Drawing the
	// parent as well says nothing more and, on a map where everything shares
	// one parent, is most of the ink.
	const blocked = new Set(full.edges.filter(edge => edge.kind === LINK_KIND.BLOCKS).map(edge => edge.to))

	const graph = {
		...full,
		edges: full.edges.filter(edge => edge.kind !== LINK_KIND.PARENT || !blocked.has(edge.to)),
	}

	const { boxes, lines, width, height } = place(graph)

	if (boxes.length === 0) return null

	const lit = focus === undefined ? undefined : connectedTo(graph.edges, focus)
	const dims = (id: IssueId) => lit !== undefined && !lit.has(id)

	return (
		<div className='-mx-1 overflow-x-auto px-1 pb-2'>
			<svg
				viewBox={`0 0 ${width} ${height}`}
				width={width}
				height={height}
				className='text-foreground max-w-none'
				role='img'
				aria-label='Map of this ticket and the work under it'
			>
				<defs>
					<pattern id='blocked-hatch' width={6} height={6} patternUnits='userSpaceOnUse' patternTransform='rotate(45)'>
						<rect width={6} height={6} className='fill-background' />
						<line x1={0} y1={0} x2={0} y2={6} className='stroke-border' strokeWidth={2} />
					</pattern>

					<marker id='blocks-arrow' viewBox='0 0 8 8' refX={7} refY={4} markerWidth={5} markerHeight={5} orient='auto'>
						<path d='M 0 0 L 8 4 L 0 8 z' className='fill-foreground' />
					</marker>
				</defs>

				{lines.map(line => (
					<path
						key={line.id}
						d={pathOf(line)}
						fill='none'
						className={cn(
							'transition-opacity',
							line.kind === LINK_KIND.BLOCKS ? 'stroke-foreground' : 'stroke-border',
							(dims(line.between[0]) || dims(line.between[1])) && 'opacity-20',
						)}
						strokeWidth={1}
						// A relation is a weaker claim than a dependency, so it is
						// drawn as a hint rather than a line.
						strokeDasharray={line.kind === LINK_KIND.RELATES ? '2 3' : undefined}
						markerEnd={line.kind === LINK_KIND.BLOCKS ? 'url(#blocks-arrow)' : undefined}
					/>
				))}

				{boxes.map(box => (
					<g
						key={box.id}
						className='cursor-pointer'
						onClick={() =>
							void navigate({
								to: '/$client/$domain/$slug/tickets/$ticket',
								params: { client, domain, slug: project, ticket: box.issue.slug },
							})
						}
					>
						<GraphNode issue={box.issue} x={box.x} y={box.y} dimmed={dims(box.id)} onFocus={setFocus} />
					</g>
				))}
			</svg>

			<p className='text-dimmer mt-2 text-xs'>
				Fill is the status: hatched is blocked, dashed is cancelled. An arrow points from a blocker to what it holds up;
				a dotted line is a relation. Hover a node to see only what it touches.
			</p>
		</div>
	)
}

export { WayfinderGraph }

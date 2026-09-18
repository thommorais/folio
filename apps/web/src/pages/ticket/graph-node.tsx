import { cn } from '@thom/libs/cn'
import type { Issue, IssueStatus, WayfinderType } from '_/core/domain/issue'
import { NODE } from './graph-geometry'

// The wayfinder type is carried by the outline's shape rather than a colour, so
// the graph still reads in either theme and without colour vision. Status is
// the fill on top of it.
const SHAPES: Record<WayfinderType | 'none', (w: number, h: number) => string> = {
	// A diamond: the map everything else hangs off.
	map: (w, h) => `M ${w / 2} 0 L ${w} ${h / 2} L ${w / 2} ${h} L 0 ${h / 2} Z`,
	// A hexagon, like a question with two sides.
	grilling: (w, h) =>
		`M ${h / 2} 0 L ${w - h / 2} 0 L ${w} ${h / 2} L ${w - h / 2} ${h} L ${h / 2} ${h} L 0 ${h / 2} Z`,
	// A stadium: open at both ends, like an enquiry.
	research: (w, h) =>
		`M ${h / 2} 0 L ${w - h / 2} 0 A ${h / 2} ${h / 2} 0 0 1 ${w - h / 2} ${h} L ${h / 2} ${h} A ${h / 2} ${h / 2} 0 0 1 ${h / 2} 0 Z`,
	// A cut corner, marking something built to be thrown away.
	prototype: (w, h) => `M 0 0 L ${w - 12} 0 L ${w} 12 L ${w} ${h} L 0 ${h} Z`,
	task: (w, h) => `M 0 0 L ${w} 0 L ${w} ${h} L 0 ${h} Z`,
	none: (w, h) => `M 0 0 L ${w} 0 L ${w} ${h} L 0 ${h} Z`,
}

// How much of the box the outline actually leaves for text at mid height. A
// diamond pinches hardest, so it gives up the most.
const TEXT_INSET: Record<WayfinderType | 'none', number> = {
	map: NODE.width - 56,
	grilling: NODE.width - 44,
	research: NODE.width - 36,
	prototype: NODE.width - 32,
	task: NODE.width - 20,
	none: NODE.width - 20,
}

const textWidth = (wayfinder: WayfinderType | undefined): number => TEXT_INSET[wayfinder ?? 'none']

// Status is the fill, so a glance says how far along the row is without
// reading it: outline open, hatched blocked, solid done.
const fillFor = (status: IssueStatus, blocked: boolean): string => {
	if (status === 'done') return 'fill-foreground/15'
	if (status === 'cancelled') return 'fill-transparent'
	if (blocked || status === 'blocked') return 'fill-[url(#blocked-hatch)]'
	if (status === 'in_progress') return 'fill-foreground/8'
	return 'fill-background'
}

type Props = {
	readonly issue: Issue
	readonly x: number
	readonly y: number
	readonly dimmed: boolean
	readonly onFocus: (id: Issue['id'] | undefined) => void
}

const GraphNode = ({ issue, x, y, dimmed, onFocus }: Props) => {
	const shape = SHAPES[issue.wayfinder ?? 'none'](NODE.width, NODE.height)
	const terminal = issue.status === 'done' || issue.status === 'cancelled'

	return (
		<g
			transform={`translate(${x} ${y})`}
			className={cn('transition-opacity', dimmed && 'opacity-25')}
			onMouseEnter={() => onFocus(issue.id)}
			onMouseLeave={() => onFocus(undefined)}
		>
			<path
				d={shape}
				className={cn('stroke-border stroke-1', fillFor(issue.status, issue.blocked))}
				// A cancelled node keeps its outline but loses its weight.
				strokeDasharray={issue.status === 'cancelled' ? '3 3' : undefined}
			/>

			{/* A diamond and a hexagon taper, so text laid out to the full box
			    would run past the outline. Each shape declares how much of its
			    width actually holds text. */}
			<foreignObject
				x={(NODE.width - textWidth(issue.wayfinder)) / 2}
				y={8}
				width={textWidth(issue.wayfinder)}
				height={NODE.height - 16}
			>
				<div className='flex h-full flex-col justify-center gap-0.5 overflow-hidden'>
					<span
						className={cn(
							'truncate text-xs leading-tight',
							terminal ? 'text-dim' : 'text-foreground',
							issue.status === 'cancelled' && 'line-through',
						)}
					>
						{issue.title}
					</span>
					<span className='text-dimmer truncate text-[10px]'>
						{issue.wayfinder ?? 'ticket'}
						{issue.dependsOn.length > 0 && ` · waits on ${issue.dependsOn.length}`}
					</span>
				</div>
			</foreignObject>
		</g>
	)
}

export { GraphNode }

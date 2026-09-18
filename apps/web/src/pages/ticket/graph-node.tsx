import { cn } from '@thom/libs/cn'
import type { Issue, IssueStatus } from '_/core/domain/issue'
import { NODE } from './graph-geometry'

// Status is the fill, so a glance says how far along the row is without
// reading it: outline open, hatched blocked, solid done. The wayfinder type is
// named on the node rather than drawn, since a shape per type cost more width
// than it earned: the tapering ones had to be drawn far wider than their box
// just to hold a title.
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
	const terminal = issue.status === 'done' || issue.status === 'cancelled'

	return (
		<g
			transform={`translate(${x} ${y})`}
			className={cn('transition-opacity', dimmed && 'opacity-25')}
			onMouseEnter={() => onFocus(issue.id)}
			onMouseLeave={() => onFocus(undefined)}
		>
			<rect
				width={NODE.width}
				height={NODE.height}
				className={cn('stroke-border stroke-1', fillFor(issue.status, issue.blocked))}
				// A cancelled node keeps its outline but loses its weight.
				strokeDasharray={issue.status === 'cancelled' ? '3 3' : undefined}
			/>

			<foreignObject x={10} y={8} width={NODE.width - 20} height={NODE.height - 16}>
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

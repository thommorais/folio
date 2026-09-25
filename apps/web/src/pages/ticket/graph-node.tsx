import { cn } from '@thom/libs/cn'
import { ISSUE_STATUS, isTerminal, type Issue, type IssueStatus } from '_/core/domain/issue'
import { NODE } from './graph-geometry'

// Status is the fill, so a glance says how far along the row is without
// reading it: outline open, hatched blocked, solid done. The wayfinder type is
// named on the node rather than drawn, since a shape per type cost more width
// than it earned: the tapering ones had to be drawn far wider than their box
// just to hold a title.
const fillFor = (status: IssueStatus, blocked: boolean): string => {
	if (status === ISSUE_STATUS.DONE) return 'fill-foreground/15'
	if (status === ISSUE_STATUS.CANCELLED) return 'fill-transparent'
	if (blocked || status === ISSUE_STATUS.BLOCKED) return 'fill-[url(#blocked-hatch)]'
	if (status === ISSUE_STATUS.IN_PROGRESS) return 'fill-foreground/8'
	return 'fill-background'
}

type Props = {
	readonly issue: Issue
	readonly x: number
	readonly y: number
	readonly dimmed: boolean
	readonly waiting: number
	readonly onFocus: (id: Issue['id'] | undefined) => void
}

const GraphNode = ({ issue, x, y, dimmed, waiting, onFocus }: Props) => {
	const terminal = isTerminal(issue.status)

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
				strokeDasharray={issue.status === ISSUE_STATUS.CANCELLED ? '3 3' : undefined}
			/>

			<foreignObject x={10} y={8} width={NODE.width - 20} height={NODE.height - 16}>
				<div className='flex h-full flex-col justify-center gap-0.5 overflow-hidden'>
					<span
						className={cn(
							'truncate text-xs leading-tight',
							terminal ? 'text-dim' : 'text-foreground',
							issue.status === ISSUE_STATUS.CANCELLED && 'line-through',
						)}
					>
						{issue.title}
					</span>
					<span className='text-dimmer truncate text-[10px]'>
						{issue.wayfinder ?? 'ticket'}
						{waiting > 0 && ` · waits on ${waiting}`}
					</span>
				</div>
			</foreignObject>
		</g>
	)
}

export { GraphNode }

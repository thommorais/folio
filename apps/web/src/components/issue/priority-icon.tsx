import { cn } from '@thom/libs/cn'
import { PRIORITY, type Priority } from '_/core/domain/issue'

const FILLED: Record<Priority, number> = { [PRIORITY.LOW]: 1, [PRIORITY.MEDIUM]: 2, [PRIORITY.HIGH]: 3 }

const LABEL: Record<Priority, string> = {
	[PRIORITY.LOW]: 'Low priority',
	[PRIORITY.MEDIUM]: 'Medium priority',
	[PRIORITY.HIGH]: 'High priority',
}

export const PriorityIcon = ({ priority, className }: { readonly priority: Priority; readonly className?: string }) => (
	<svg viewBox='0 0 14 14' role='img' aria-label={LABEL[priority]} className={cn('text-dim size-3.5 shrink-0', className)}>
		<title>{LABEL[priority]}</title>
		{[0, 1, 2].map(bar => (
			<rect
				key={bar}
				x={1.5 + bar * 4}
				y={9 - bar * 3}
				width='3'
				height={3.5 + bar * 3}
				rx='0.75'
				fill='currentColor'
				opacity={bar < FILLED[priority] ? 1 : 0.25}
			/>
		))}
	</svg>
)

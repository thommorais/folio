import { cn } from '@thom/libs/cn'
import { ISSUE_STATUS, type IssueStatus } from '_/core/domain/issue'
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels'

const TONE: Record<IssueStatus, string> = {
	[ISSUE_STATUS.OPEN]: 'text-dim',
	[ISSUE_STATUS.IN_PROGRESS]: 'text-amber-500',
	[ISSUE_STATUS.BLOCKED]: 'text-destructive',
	[ISSUE_STATUS.DONE]: 'text-indigo-500',
	[ISSUE_STATUS.CANCELLED]: 'text-dimmer',
}

const Glyph = ({ status }: { readonly status: IssueStatus }) => {
	switch (status) {
		case ISSUE_STATUS.OPEN:
			return <circle cx='7' cy='7' r='5.5' stroke='currentColor' strokeWidth='1.5' strokeDasharray='2.2 1.6' />
		case ISSUE_STATUS.IN_PROGRESS:
			return (
				<>
					<circle cx='7' cy='7' r='5.5' stroke='currentColor' strokeWidth='1.5' />
					<path d='M7 3.5a3.5 3.5 0 0 1 0 7z' fill='currentColor' />
				</>
			)
		case ISSUE_STATUS.BLOCKED:
			return (
				<>
					<circle cx='7' cy='7' r='5.5' stroke='currentColor' strokeWidth='1.5' />
					<path d='M7 4.2v3.4' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round' />
					<circle cx='7' cy='9.8' r='0.85' fill='currentColor' />
				</>
			)
		case ISSUE_STATUS.DONE:
			return (
				<>
					<circle cx='7' cy='7' r='6.25' fill='currentColor' />
					<path
						d='M4.4 7.1l1.8 1.8 3.4-3.6'
						stroke='var(--color-background)'
						strokeWidth='1.5'
						strokeLinecap='round'
						strokeLinejoin='round'
					/>
				</>
			)
		case ISSUE_STATUS.CANCELLED:
			return (
				<>
					<circle cx='7' cy='7' r='6.25' fill='currentColor' />
					<path d='M5 5l4 4M9 5l-4 4' stroke='var(--color-background)' strokeWidth='1.5' strokeLinecap='round' />
				</>
			)
	}
}

export const StatusIcon = ({ status, className }: { readonly status: IssueStatus; readonly className?: string }) => (
	<svg
		viewBox='0 0 14 14'
		fill='none'
		role='img'
		aria-label={ISSUE_STATUS_LABELS[status]}
		className={cn('size-3.5 shrink-0', TONE[status], className)}
	>
		<title>{ISSUE_STATUS_LABELS[status]}</title>
		<Glyph status={status} />
	</svg>
)

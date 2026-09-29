import { cn } from '@thom/libs/cn';
import { ISSUE_STATUS, isTerminal, type Issue } from '_/core/domain/issue';
import type { ReactNode } from 'react';
import { PriorityIcon } from './priority-icon';
import { StatusIcon } from './status-icon';
import { Tags } from './tag';

const dayMonth = new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short' })

type Props = {
	readonly issue: Issue
	readonly title?: ReactNode
	readonly muted?: boolean
	readonly strike?: boolean
	readonly extra?: ReactNode
}

export const IssueLine = ({ issue, title, muted = isTerminal(issue.status), strike = false, extra }: Props) => (
	<div className='flex h-5 items-center gap-3'>
		<PriorityIcon priority={issue.priority} className={cn(muted && 'opacity-50')} />
		<StatusIcon status={issue.status} />

		<span
			className={cn(
				'min-w-0 flex-1 truncate text-sm',
				muted && 'text-dim',
				strike && issue.status === ISSUE_STATUS.DONE && 'line-through',
			)}
		>
			{title ?? issue.title}
		</span>

		<span className='hidden shrink-0 items-center gap-3 sm:flex'>
			<Tags tags={issue.tags} className='flex-nowrap' />
			{extra}
		</span>

		<time dateTime={issue.updatedAt.toISOString()} className='text-dimmer w-12 shrink-0 text-right text-xs tabular-nums'>
			{dayMonth.format(issue.updatedAt)}
		</time>
	</div>
)

export const ISSUE_LINE_INSET = 'pl-[3.25rem]'

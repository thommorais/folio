import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import { StaggerItem } from '_/components/motion/stagger'
import { Flash } from '_/components/motion/flash'
import { EmptyState } from '_/components/empty-state'
import { LoadError } from '_/components/load-error'

export type WorkItem = {
	readonly id: string
	readonly to: string
	readonly params: Record<string, string>
	readonly line: ReactNode
	readonly changedAt: number
}

type Props = {
	readonly title: string
	readonly items: readonly WorkItem[] | undefined
	readonly message: string | undefined
	readonly emptyLabel: string
	readonly command: string
	readonly filtered: boolean
	readonly to: string
	readonly params: Record<string, string>
}

export const Section = ({ title, items, message, emptyLabel, command, filtered, to, params }: Props) => (
	<section className='space-y-2'>
		<header className='flex items-baseline gap-2'>
			<Link
				to={to}
				params={params}
				className='text-dim hover:text-foreground text-xs tracking-widest uppercase transition-colors'
			>
				{title}
			</Link>
			{items !== undefined && <span className='text-dimmer text-xs tabular-nums'>{items.length}</span>}
		</header>

		{message !== undefined && <LoadError message={message} />}

		{items !== undefined && items.length === 0 && filtered && <p className='text-dim text-sm'>No {emptyLabel} match.</p>}

		{items !== undefined && items.length === 0 && !filtered && (
			<EmptyState message={`No ${emptyLabel} yet.`} command={command} />
		)}

		{items !== undefined && items.length > 0 && (
			<ul className='border-border divide-border divide-y border'>
				{items.map((item, index) => (
					<StaggerItem key={item.id} index={index} as='li'>
						<Link
							to={item.to}
							params={item.params}
							data-row
							className='hover:bg-accent/40 active:bg-accent/60 relative block px-4 py-2.5 transition-colors'
						>
							<Flash on={item.changedAt} />
							{item.line}
						</Link>
					</StaggerItem>
				))}
			</ul>
		)}
	</section>
)

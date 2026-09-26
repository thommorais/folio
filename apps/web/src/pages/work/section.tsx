import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'

export type WorkItem = {
	readonly id: string
	readonly to: string
	readonly params: Record<string, string>
	readonly line: ReactNode
}

type Props = {
	readonly title: string
	readonly items: readonly WorkItem[] | undefined
	readonly message: string | undefined
	readonly emptyLabel: string
	readonly filtered: boolean
	readonly to: string
	readonly params: Record<string, string>
}

export const Section = ({ title, items, message, emptyLabel, filtered, to, params }: Props) => (
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

		{message !== undefined && <p className='text-destructive text-sm'>{message}</p>}

		{message === undefined && items === undefined && (
			<div className='border-border divide-border divide-y border'>
				{[0, 1].map(key => (
					<Skeleton key={key} className='h-10' />
				))}
			</div>
		)}

		{items !== undefined && items.length === 0 && (
			<p className='text-dim text-sm'>{filtered ? `No ${emptyLabel} match.` : `No ${emptyLabel} yet.`}</p>
		)}

		{items !== undefined && items.length > 0 && (
			<ul className='border-border divide-border divide-y border'>
				{items.map((item, index) => (
					<StaggerItem key={item.id} index={index} as='li'>
						<Link
							to={item.to}
							params={item.params}
							className='hover:bg-accent/40 block px-4 py-2.5 transition-colors'
						>
							{item.line}
						</Link>
					</StaggerItem>
				))}
			</ul>
		)}
	</section>
)

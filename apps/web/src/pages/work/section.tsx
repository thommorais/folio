import { Link } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'

export type WorkItem = {
	readonly id: string
	readonly title: string
	readonly status: string
	readonly muted: boolean
	// Todos get a checkbox; a ticket or a plan is not something you tick off.
	readonly checked?: boolean
	readonly meta: readonly string[]
	readonly to: string
	readonly params: Record<string, string>
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
					<Skeleton key={key} className='h-11' />
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
							className='hover:bg-accent/40 flex items-center gap-3 px-4 py-3 transition-colors'
						>
							{item.checked !== undefined && (
								<span
									className={cn(
										'border-border size-4 shrink-0 border',
										item.checked && 'bg-foreground border-foreground',
									)}
								/>
							)}

							<span className={cn('min-w-0 flex-1 truncate text-sm', item.muted && 'text-dim line-through')}>
								{item.title}
							</span>

							<span className='hidden shrink-0 items-center gap-2 sm:flex'>
								{item.meta.map(entry => (
									<Badge key={entry} color='muted'>
										{entry}
									</Badge>
								))}
							</span>

							<span className='text-dim w-24 shrink-0 text-right text-xs'>{item.status}</span>
						</Link>
					</StaggerItem>
				))}
			</ul>
		)}
	</section>
)

import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import { StaggerItem } from '_/components/motion/stagger'
import { Flash } from '_/components/motion/flash'
import { EmptyState } from '_/components/empty-state'
import { LoadError } from '_/components/load-error'
import { GroupHeader } from '_/components/group/group-header'
import type { GroupTarget } from '_/core/domain/grouping'

export type WorkItem = {
	readonly id: string
	readonly to: string
	readonly params: Record<string, string>
	readonly line: ReactNode
	readonly changedAt: number
}

export type WorkGroup = {
	readonly target: GroupTarget | undefined
	readonly items: readonly WorkItem[]
}

type Props = {
	readonly title: string
	readonly items: readonly WorkItem[] | undefined
	readonly groups?: readonly WorkGroup[]
	readonly slug?: string
	readonly message: string | undefined
	readonly emptyLabel: string
	readonly command: string
	readonly filtered: boolean
	readonly to: string
	readonly params: Record<string, string>
}

const ItemList = ({ items }: { readonly items: readonly WorkItem[] }) => (
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
)

export const Section = ({ title, items, groups, slug, message, emptyLabel, command, filtered, to, params }: Props) => (
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

		{items !== undefined && items.length > 0 && (groups === undefined || slug === undefined) && <ItemList items={items} />}

		{items !== undefined && items.length > 0 && groups !== undefined && slug !== undefined && (
			<div className='space-y-4'>
				{groups.map(({ target, items: grouped }) => (
					<div key={target ? `${target.kind}:${target.id}` : 'none'} className='space-y-2'>
						{(target !== undefined || groups.length > 1) && <GroupHeader target={target} slug={slug} />}
						<ItemList items={grouped} />
					</div>
				))}
			</div>
		)}
	</section>
)

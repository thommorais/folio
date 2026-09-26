import { Link } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Card, CardHeader } from '@thom/ui/card'
import { AnimatedNumber } from '_/components/motion/animated-number'
import { StaggerItem } from '_/components/motion/stagger'
import { useCounts } from '_/app/use-counts'
import { entities, type Entity } from '_/app/counts'
import { Status } from '_/lib/async-status'
import { LoadError } from '_/components/load-error'

const labels: Record<Entity, string> = {
	tickets: 'Tickets',
	plans: 'Plans',
	todos: 'Todos',
	journal: 'Journal',
}

const routes: Record<Entity, string> = {
	tickets: '/$client/$domain/$slug/tickets',
	plans: '/$client/$domain/$slug/plans',
	todos: '/$client/$domain/$slug/todos',
	journal: '/$client/$domain/$slug/journal',
}

type Props = {
	readonly client: string
	readonly domain: string
	readonly slug: string
}

export const ProjectCounts = ({ client, domain, slug }: Props) => {
	const state = useCounts(slug)

	if (state.status === Status.Failed) {
		return <LoadError message={state.message} />
	}

	return (
		<div className='grid grid-cols-4 gap-2 sm:gap-4'>
			{entities.map((entity, index) => (
				<StaggerItem key={entity} index={index}>
					<Link to={routes[entity]} params={{ client, domain, slug }} className='block'>
						<Card interactive>
							<CardHeader className='space-y-0.5 p-3 sm:space-y-1.5 sm:p-6'>
								<span className='text-dim truncate text-[11px] sm:text-xs'>{labels[entity]}</span>
								<span className={cn('font-serif text-xl tabular-nums sm:text-2xl', state.status === Status.Loading && 'invisible')}>
									{state.status === Status.Ready ? <AnimatedNumber value={state.counts[entity]} /> : 0}
								</span>
							</CardHeader>
						</Card>
					</Link>
				</StaggerItem>
			))}
		</div>
	)
}

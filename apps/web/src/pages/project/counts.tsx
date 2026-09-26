import { Link } from '@tanstack/react-router'
import { Card, CardHeader } from '@thom/ui/card'
import { AnimatedNumber } from '_/components/motion/animated-number'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useCounts } from '_/app/use-counts'
import { entities, type Entity } from '_/app/counts'
import { Status } from '_/lib/async-status'

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
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return (
		<div className='grid grid-cols-4 gap-2 sm:gap-4'>
			{entities.map((entity, index) => (
				<StaggerItem key={entity} index={index}>
					<Link to={routes[entity]} params={{ client, domain, slug }} className='block'>
						<Card interactive>
							<CardHeader className='space-y-0.5 p-3 sm:space-y-1.5 sm:p-6'>
								<span className='text-dim truncate text-[11px] sm:text-xs'>{labels[entity]}</span>
								{state.status === Status.Loading ? (
									<Skeleton className='mt-1 h-6 w-8 sm:h-8 sm:w-10' />
								) : (
									<span className='font-serif text-xl tabular-nums sm:text-2xl'>
										<AnimatedNumber value={state.counts[entity]} />
									</span>
								)}
							</CardHeader>
						</Card>
					</Link>
				</StaggerItem>
			))}
		</div>
	)
}

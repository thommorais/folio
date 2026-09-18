import { Link } from '@tanstack/react-router'
import { Card, CardHeader } from '@thom/ui/card'
import { AnimatedNumber } from '_/components/motion/animated-number'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useCounts } from '_/app/use-counts'
import { entities, type Entity } from '_/app/counts'

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

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return (
		<div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
			{entities.map((entity, index) => (
				<StaggerItem key={entity} index={index}>
					<Link to={routes[entity]} params={{ client, domain, slug }} className='block'>
						<Card interactive>
							<CardHeader>
								<span className='text-dim text-xs'>{labels[entity]}</span>
								{state.status === 'loading' ? (
									<Skeleton className='mt-1 h-8 w-10' />
								) : (
									<span className='font-serif text-2xl tabular-nums'>
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

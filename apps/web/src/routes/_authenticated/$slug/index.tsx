import { createFileRoute, useParams } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
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
	docs: 'Docs',
}

const Counts = ({ slug }: { readonly slug: string }) => {
	const state = useCounts(slug)

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return (
		<div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-5'>
			{entities.map((entity, index) => (
				<StaggerItem key={entity} index={index}>
					<Card>
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
				</StaggerItem>
			))}
		</div>
	)
}

const Overview = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/' })

	return (
		<div className='space-y-6'>
			<Counts slug={slug} />

			<Card>
				<CardHeader>
					<CardTitle>Active plan</CardTitle>
					<CardDescription>Ship full text search</CardDescription>

					<div className='pt-4'>
						<div className='bg-accent h-1 w-full'>
							<motion.div
								className='bg-foreground h-1 origin-left'
								initial={{ scaleX: 0 }}
								animate={{ scaleX: 0.4 }}
								transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1], delay: 0.15 }}
								style={{ width: '100%' }}
							/>
						</div>
						<span className='text-dimmer pt-2 text-xs'>2 of 5 done</span>
					</div>
				</CardHeader>
			</Card>
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/$slug/')({
	component: Overview,
})

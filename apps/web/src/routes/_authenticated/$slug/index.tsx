import { createFileRoute, useParams } from '@tanstack/react-router'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
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
			{entities.map(entity => (
				<Card key={entity}>
					<CardHeader>
						<span className='text-dim text-xs'>{labels[entity]}</span>
						{state.status === 'loading' ? (
							<span className='bg-accent/40 mt-1 h-8 w-10 animate-pulse' />
						) : (
							<span className='font-serif text-2xl'>{state.counts[entity]}</span>
						)}
					</CardHeader>
				</Card>
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
							<div className='bg-foreground h-1' style={{ width: '40%' }} />
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

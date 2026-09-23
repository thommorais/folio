import { createFileRoute, Link, useParams } from '@tanstack/react-router'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useClient } from '_/app/use-client'
import { useDomains } from '_/app/use-domains'

const ClientPage = () => {
	const { client: slug } = useParams({ from: '/_authenticated/$client/' })
	const client = useClient(slug)
	const state = useDomains(
		client.status === 'ready' ? { clientId: client.client.id } : undefined,
		client.status !== 'ready',
	)

	if (client.status === 'failed') return <p className='text-destructive text-sm'>{client.message}</p>

	return (
		<div className='space-y-6'>
			<div className='space-y-1'>
				<h1 className='font-serif text-3xl'>{client.status === 'ready' ? client.client.name : slug}</h1>
				{client.status === 'ready' && client.client.descr && <p className='text-dim text-sm'>{client.client.descr}</p>}
			</div>

			{state.status === 'loading' && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='border-border h-[124px] border' />
					))}
				</div>
			)}

			{state.status === 'failed' && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === 'ready' && state.domains.length === 0 && (
				<p className='text-dim text-sm'>No domains in this client yet.</p>
			)}

			{state.status === 'ready' && state.domains.length > 0 && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{state.domains.map((domain, index) => (
						<StaggerItem key={domain.id} index={index}>
							<Link to='/$client/$domain' params={{ client: slug, domain: domain.slug }} className='block'>
								<Card interactive>
									<CardHeader>
										<CardTitle>{domain.name}</CardTitle>

										<CardDescription>{domain.descr || 'No description.'}</CardDescription>

										<div className='text-dimmer flex items-center gap-2 pt-2 text-xs'>
											<span>{domain.slug}</span>
											<span>&middot;</span>
											<span>
												{domain.members.length} {domain.members.length === 1 ? 'member' : 'members'}
											</span>
										</div>
									</CardHeader>
								</Card>
							</Link>
						</StaggerItem>
					))}
				</div>
			)}
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/$client/')({
	component: ClientPage,
})

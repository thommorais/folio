import { createFileRoute, Link } from '@tanstack/react-router'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useClients } from '_/app/use-clients'
import { useDomains } from '_/app/use-domains'
import { buildInsights } from '_/pages/clients/insights'
import { SummaryTicker } from '_/pages/clients/summary-ticker'
import { WelcomeGreeting } from '_/pages/clients/welcome'
import { Status } from '_/lib/async-status'

const Clients = () => {
	const state = useClients()
	const domainState = useDomains()

	const domains = domainState.status === Status.Ready ? domainState.domains : []

	return (
		<div className='mx-auto w-full max-w-5xl space-y-8'>
			<div className='flex w-full flex-col items-center pt-6 pb-4 text-center'>
				<WelcomeGreeting />
				{state.status === Status.Ready && <SummaryTicker insights={buildInsights(state.clients, domains)} />}
			</div>

			{state.status === Status.Loading && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='border-border h-[124px] border' />
					))}
				</div>
			)}

			{state.status === Status.Failed && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === Status.Ready && state.clients.length === 0 && (
				<p className='text-dim text-sm'>No clients yet.</p>
			)}

			{state.status === Status.Ready && state.clients.length > 0 && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{state.clients.map((client, index) => {
						const count = domains.filter(domain => domain.clientId === client.id).length

						return (
							<StaggerItem key={client.id} index={index}>
								<Link to='/$client' params={{ client: client.slug }} className='block'>
									<Card interactive>
										<CardHeader>
											<CardTitle>{client.name}</CardTitle>

											<CardDescription>{client.descr || 'No description.'}</CardDescription>

											<div className='text-dimmer flex items-center gap-2 pt-2 text-xs'>
												<span>{client.slug}</span>
												<span>&middot;</span>
												<span>
													{count} {count === 1 ? 'domain' : 'domains'}
												</span>
											</div>
										</CardHeader>
									</Card>
								</Link>
							</StaggerItem>
						)
					})}
				</div>
			)}
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/')({
	component: Clients,
})

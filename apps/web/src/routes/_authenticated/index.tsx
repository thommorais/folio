import { createFileRoute, Link, redirect } from '@tanstack/react-router'
import { lastProject } from '_/adapters/browser/last-project'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { StaggerItem } from '_/components/motion/stagger'
import { useClients } from '_/app/use-clients'
import { useDomains } from '_/app/use-domains'
import { buildInsights } from '_/pages/clients/insights'
import { SummaryTicker } from '_/pages/clients/summary-ticker'
import { WelcomeGreeting } from '_/pages/clients/welcome'
import { Status } from '_/lib/async-status'
import { LoadError } from '_/components/load-error'

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

			{state.status === Status.Failed && <LoadError message={state.message} />}

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
	beforeLoad: () => {
		const last = lastProject.resume()
		if (last) throw redirect({ to: '/$client/$domain/$slug', params: last })
	},
	component: Clients,
})

import { createFileRoute, Link, useParams } from '@tanstack/react-router'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { StaggerItem } from '_/components/motion/stagger'
import { useClient } from '_/app/use-client'
import { RecordGone } from '_/components/record/record-gone'
import { useDomains } from '_/app/use-domains'
import { Status } from '_/lib/async-status'
import { LoadError } from '_/components/load-error'

const ClientPage = () => {
	const { client: slug } = useParams({ from: '/_authenticated/$client/' })
	const client = useClient(slug)
	const state = useDomains(
		client.status === Status.Ready ? { clientId: client.client.id } : undefined,
		client.status !== Status.Ready,
	)

	if (client.status === Status.Failed) return <LoadError message={client.message} />
	if (client.status === Status.Gone) {
		return (
			<RecordGone title={client.title}>
				<Link to='/' className='text-sm underline'>
					Back to clients
				</Link>
			</RecordGone>
		)
	}

	return (
		<div className='space-y-6'>
			<div className='space-y-1'>
				<h1 className='font-serif text-3xl'>{client.status === Status.Ready ? client.client.name : slug}</h1>
				{client.status === Status.Ready && client.client.descr && (
					<p className='text-dim text-sm'>{client.client.descr}</p>
				)}
			</div>

			{state.status === Status.Failed && <LoadError message={state.message} />}

			{state.status === Status.Ready && state.domains.length === 0 && (
				<p className='text-dim text-sm'>No domains in this client yet.</p>
			)}

			{state.status === Status.Ready && state.domains.length > 0 && (
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

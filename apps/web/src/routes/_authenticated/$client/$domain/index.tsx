import { createFileRoute, Link, useParams } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { useDomain } from '_/app/use-domain'
import { useProjects } from '_/app/use-projects'
import { MemberRoster } from '_/pages/domain/member-roster'
import { Status } from '_/lib/async-status'

const DomainPage = () => {
	const { client, domain: slug } = useParams({ from: '/_authenticated/$client/$domain/' })
	const domain = useDomain(client, slug)
	const state = useProjects(
		domain.status === Status.Ready ? { domainId: domain.domain.id } : undefined,
		domain.status !== Status.Ready,
	)

	if (domain.status === Status.Failed) return <p className='text-destructive text-sm'>{domain.message}</p>

	return (
		<div className='space-y-6'>
			<div className='space-y-1'>
				<h1 className='font-serif text-3xl'>{domain.status === Status.Ready ? domain.domain.name : slug}</h1>
				{domain.status === Status.Ready && domain.domain.descr && (
					<p className='text-dim text-sm'>{domain.domain.descr}</p>
				)}
			</div>

			{domain.status === Status.Ready && <MemberRoster members={domain.domain.members} />}

			{state.status === Status.Loading && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='border-border h-[124px] border' />
					))}
				</div>
			)}

			{state.status === Status.Failed && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === Status.Ready && state.projects.length === 0 && (
				<p className='text-dim text-sm'>No projects in this domain yet.</p>
			)}

			{state.status === Status.Ready && state.projects.length > 0 && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{state.projects.map((project, index) => (
						<StaggerItem key={project.id} index={index}>
							<Link to='/$client/$domain/$slug' params={{ client, domain: slug, slug: project.slug }} className='block'>
								<Card interactive>
									<CardHeader>
										<div className='flex items-start justify-between gap-4'>
											<CardTitle>{project.name}</CardTitle>
											{project.archived && <Badge color='muted'>Archived</Badge>}
										</div>

										<CardDescription>{project.descr || 'No description.'}</CardDescription>

										<div className='text-dimmer flex items-center gap-2 pt-2 text-xs'>
											<span>{project.slug}</span>
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

export const Route = createFileRoute('/_authenticated/$client/$domain/')({
	component: DomainPage,
})

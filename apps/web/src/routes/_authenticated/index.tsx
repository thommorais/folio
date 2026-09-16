import { createFileRoute, Link } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { Card, CardDescription, CardHeader, CardTitle } from '@thom/ui/card'
import { useProjects } from '_/app/use-projects'
import { buildInsights } from '_/pages/projects/insights'
import { SummaryTicker } from '_/pages/projects/summary-ticker'
import { WelcomeGreeting } from '_/pages/projects/welcome'

const Projects = () => {
	const state = useProjects()

	return (
		<div className='mx-auto w-full max-w-5xl space-y-8'>
			<div className='flex w-full flex-col items-center pt-6 pb-4 text-center'>
				<WelcomeGreeting />
				{state.status === 'ready' && <SummaryTicker insights={buildInsights(state.projects)} />}
			</div>

			{state.status === 'loading' && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{[0, 1].map(key => (
						<div key={key} className='border-border bg-accent/40 h-[124px] animate-pulse border' />
					))}
				</div>
			)}

			{state.status === 'failed' && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === 'ready' && state.projects.length === 0 && <p className='text-dim text-sm'>No projects yet.</p>}

			{state.status === 'ready' && state.projects.length > 0 && (
				<div className='grid gap-4 sm:grid-cols-2'>
					{state.projects.map(project => (
						<Link key={project.id} to='/$slug' params={{ slug: project.slug }} className='block'>
							<Card interactive>
								<CardHeader>
									<div className='flex items-start justify-between gap-4'>
										<CardTitle>{project.name}</CardTitle>
										{project.archived && <Badge color='muted'>Archived</Badge>}
									</div>

									<CardDescription>{project.descr || 'No description.'}</CardDescription>

									<div className='text-dimmer flex items-center gap-2 pt-2 text-xs'>
										<span>{project.slug}</span>
										<span>&middot;</span>
										<span>
											{project.members.length} {project.members.length === 1 ? 'member' : 'members'}
										</span>
									</div>
								</CardHeader>
							</Card>
						</Link>
					))}
				</div>
			)}
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/')({
	component: Projects,
})

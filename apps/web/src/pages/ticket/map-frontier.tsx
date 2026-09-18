import { Link } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { useIssues } from '_/app/use-issues'
import { partitionChildren } from '_/core/domain/frontier'
import type { Issue } from '_/core/domain/issue'
import { useScope } from '_/routing/use-scope'
import { WayfinderGraph } from './wayfinder-graph'

type Props = {
	readonly project: string
	readonly map: Issue
}

const Group = ({
	title,
	tickets,
	project,
}: {
	readonly title: string
	readonly tickets: readonly Issue[]
	readonly project: string
}) => {
	const { client, domain } = useScope()

	if (tickets.length === 0) return null

	return (
		<div className='space-y-2'>
			<h3 className='text-dimmer text-xs tracking-wide uppercase'>
				{title} ({tickets.length})
			</h3>
			<ul className='border-border divide-border divide-y border'>
				{tickets.map(child => (
					<li key={child.id}>
						<Link
							to='/$client/$domain/$slug/tickets/$ticket'
							params={{ client, domain, slug: project, ticket: child.slug }}
							className='hover:bg-accent/40 flex items-center gap-3 px-4 py-3 transition-colors'
						>
							<span className='flex-1 truncate text-sm'>{child.title}</span>
							{child.wayfinder && <Badge color='muted'>{child.wayfinder}</Badge>}
							{child.dependsOn.length > 0 && (
								<span className='text-dimmer shrink-0 text-xs'>waits on {child.dependsOn.length}</span>
							)}
						</Link>
					</li>
				))}
			</ul>
		</div>
	)
}

const MapFrontier = ({ project, map }: Props) => {
	// A map's decisions are tickets. Todos filed under the same ticket are its
	// steps, not places the work can go next, and they have their own section.
	const children = useIssues(project, { kind: 'ticket', parentId: map.id })

	if (children.status !== 'ready') return null

	if (children.issues.length === 0) {
		return <p className='text-dim text-sm'>No decision tickets on this map yet.</p>
	}

	const { frontier, blocked, claimed, done } = partitionChildren(children.issues)

	return (
		<section className='space-y-4'>
			<h2 className='text-foreground text-sm font-medium'>The map</h2>

			{/* The graph answers how the work hangs together; the groups below
			    answer what to pick up next. Neither replaces the other. */}
			<WayfinderGraph project={project} mapId={map.id} issues={[map, ...children.issues]} />

			<Group title='Takeable' tickets={frontier} project={project} />
			<Group title='Blocked' tickets={blocked} project={project} />
			<Group title='Claimed' tickets={claimed} project={project} />
			<Group title='Done' tickets={done} project={project} />
		</section>
	)
}

export { MapFrontier }

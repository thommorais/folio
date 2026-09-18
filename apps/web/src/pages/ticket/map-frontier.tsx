import { Link } from '@tanstack/react-router'
import { Badge } from '@thom/ui/badge'
import { useIssues } from '_/app/use-issues'
import { partitionChildren } from '_/core/domain/frontier'
import type { Issue } from '_/core/domain/issue'

type Props = {
	readonly project: string
	readonly mapId: string
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
							to='/$slug/tickets/$ticket'
							params={{ slug: project, ticket: child.slug }}
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

const MapFrontier = ({ project, mapId }: Props) => {
	const children = useIssues(project, { parentId: mapId })

	if (children.status !== 'ready') return null

	if (children.issues.length === 0) {
		return <p className='text-dim text-sm'>No decision tickets on this map yet.</p>
	}

	const { frontier, blocked, claimed, done } = partitionChildren(children.issues)

	return (
		<section className='space-y-4'>
			<h2 className='text-foreground text-sm font-medium'>The map</h2>
			<Group title='Takeable' tickets={frontier} project={project} />
			<Group title='Blocked' tickets={blocked} project={project} />
			<Group title='Claimed' tickets={claimed} project={project} />
			<Group title='Done' tickets={done} project={project} />
		</section>
	)
}

export { MapFrontier }

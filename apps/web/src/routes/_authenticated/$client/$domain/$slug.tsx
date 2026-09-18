import { createFileRoute, Outlet, useParams } from '@tanstack/react-router'
import { ProjectTabs } from '_/components/project-tabs'
import { ProjectCounts } from '_/pages/project/counts'

const ProjectLayout = () => {
	const { client, domain, slug } = useParams({ from: '/_authenticated/$client/$domain/$slug' })

	return (
		<div className='space-y-6'>
			<ProjectCounts client={client} domain={domain} slug={slug} />

			<ProjectTabs client={client} domain={domain} slug={slug} />

			<Outlet />
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug')({
	component: ProjectLayout,
})

import { createFileRoute, Outlet, useParams } from '@tanstack/react-router'
import { useEffect } from 'react'
import { lastProject } from '_/adapters/browser/last-project'
import { ProjectCounts } from '_/pages/project/counts'

const ProjectLayout = () => {
	const { client, domain, slug } = useParams({ from: '/_authenticated/$client/$domain/$slug' })

	useEffect(() => lastProject.remember({ client, domain, slug }), [client, domain, slug])

	return (
		<div className='space-y-6'>
			<ProjectCounts client={client} domain={domain} slug={slug} />

			<Outlet />
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug')({
	component: ProjectLayout,
})

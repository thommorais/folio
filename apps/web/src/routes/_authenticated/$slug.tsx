import { createFileRoute, Outlet, useParams } from '@tanstack/react-router'
import { Breadcrumbs } from '_/components/breadcrumbs'
import { ProjectTabs } from '_/components/project-tabs'

const ProjectLayout = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug' })

	return (
		<div className='mx-auto w-full max-w-5xl space-y-6'>
			<Breadcrumbs />

			<ProjectTabs slug={slug} />

			<Outlet />
		</div>
	)
}

export const Route = createFileRoute('/_authenticated/$slug')({
	component: ProjectLayout,
})

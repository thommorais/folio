import { createFileRoute, Outlet } from '@tanstack/react-router'
import { Breadcrumbs } from '_/components/breadcrumbs'

const ClientLayout = () => (
	<div className='mx-auto w-full max-w-5xl space-y-6'>
		<Breadcrumbs />

		<Outlet />
	</div>
)

export const Route = createFileRoute('/_authenticated/$client')({
	component: ClientLayout,
})

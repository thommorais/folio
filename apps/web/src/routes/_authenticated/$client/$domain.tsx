import { createFileRoute, Outlet } from '@tanstack/react-router'

const DomainLayout = () => <Outlet />

export const Route = createFileRoute('/_authenticated/$client/$domain')({
	component: DomainLayout,
})

import { createFileRoute, Outlet } from '@tanstack/react-router'
import { AppShell } from '_/components/app-shell'
import { SignedIn } from '_/components/signed-in'

const AuthenticatedLayout = () => (
	<SignedIn>
		<AppShell>
			<Outlet />
		</AppShell>
	</SignedIn>
)

export const Route = createFileRoute('/_authenticated')({
	component: AuthenticatedLayout,
})

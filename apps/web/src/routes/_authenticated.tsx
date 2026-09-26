import { createFileRoute, Outlet } from '@tanstack/react-router'
import { useEffect } from 'react'
import { lastProject } from '_/adapters/browser/last-project'
import { AppShell } from '_/components/app-shell'
import { SignedIn } from '_/components/signed-in'

const AuthenticatedLayout = () => {
	useEffect(() => lastProject.settle(), [])

	return (
		<SignedIn>
			<AppShell>
				<Outlet />
			</AppShell>
		</SignedIn>
	)
}

export const Route = createFileRoute('/_authenticated')({
	component: AuthenticatedLayout,
})

import { createRootRoute, type ErrorComponentProps, Outlet, useRouter } from '@tanstack/react-router'
import { NuqsAdapter } from 'nuqs/adapters/tanstack-router'
import { Button } from '@thom/ui/button'
import { Toaster } from '@thom/ui/toast'
import { ContainerProvider } from '_/app/container'
import { ReloadPrompt } from '_/components/reload-prompt'

const RootComponent = () => {
	return (
		<ContainerProvider>
			<NuqsAdapter>
				<Outlet />
				<Toaster />
				<ReloadPrompt />
			</NuqsAdapter>
		</ContainerProvider>
	)
}

const ErrorComponent = ({ error }: ErrorComponentProps) => {
	const router = useRouter()

	return (
		<main className='flex min-h-dvh w-full flex-col items-center justify-center gap-6 p-8'>
			<div className='max-w-md space-y-2 text-center'>
				<h1 className='font-serif text-lg'>Something went wrong</h1>
				<p className='text-dim text-sm'>{error instanceof Error ? error.message : 'Unknown error'}</p>
			</div>

			<Button variant='outline' onClick={() => router.invalidate()}>
				Try again
			</Button>
		</main>
	)
}

export const Route = createRootRoute({
	component: RootComponent,
	errorComponent: ErrorComponent,
})

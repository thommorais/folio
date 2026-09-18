import { createFileRoute, Link, useNavigate, useRouter } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { Button } from '@thom/ui/button'
import { Input } from '@thom/ui/input'
import type { AuthError } from '_/core/ports/auth'
import { useSession, useSignIn } from '_/app/use-session'

const messageFor = (error: AuthError): string => {
	switch (error.kind) {
		case 'invalid-credentials':
			return 'Wrong email or password.'
		case 'unverified':
			return 'This account is not verified yet.'
		case 'network':
			return 'Cannot reach the server.'
		default:
			return error.message
	}
}

const Login = () => {
	const router = useRouter()
	const navigate = useNavigate()
	const signIn = useSignIn()
	const session = useSession()
	const [error, setError] = useState<string>()
	const [isSubmitting, setSubmitting] = useState(false)

	useEffect(() => {
		if (session.status === 'authenticated') {
			void navigate({ to: '/', replace: true })
		}
	}, [session.status, navigate])

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		const form = new FormData(event.currentTarget)
		setSubmitting(true)
		setError(undefined)

		const result = await signIn({
			email: String(form.get('email') ?? ''),
			password: String(form.get('password') ?? ''),
		})

		if (result.ok) {
			await router.navigate({ to: '/', replace: true })
		} else {
			setError(messageFor(result.error))
		}

		setSubmitting(false)
	}

	return (
		<main className='relative flex min-h-dvh w-full'>
			<div className='border-border bg-card hidden w-1/2 flex-col justify-between border-r p-12 lg:flex'>
				<span className='font-serif text-lg'>folio</span>
				<p className='max-w-sm font-serif text-2xl leading-snug text-balance'>
					The shared channel between developers and their coding agents.
				</p>
				<span className='text-dim text-xs'>Tickets, plans, todos and a shared journal.</span>
			</div>

			<div className='flex w-full flex-col items-center justify-center p-8 lg:w-1/2 lg:p-12'>
				<div className='flex w-full max-w-md flex-1 flex-col justify-center space-y-8'>
					<div className='space-y-2 text-center'>
						<h1 className='mb-4 font-serif text-lg lg:text-xl'>Welcome to folio</h1>
						<p className='text-dim text-sm'>Sign in to your account</p>
					</div>

					<form onSubmit={onSubmit} className='flex flex-col space-y-4'>
						<Input
							type='email'
							name='email'
							placeholder='Enter email address'
							required
							autoComplete='email'
							autoCapitalize='none'
							autoCorrect='off'
							spellCheck={false}
							aria-invalid={error ? true : undefined}
						/>

						<Input
							type='password'
							name='password'
							placeholder='Enter password'
							required
							autoComplete='current-password'
							aria-invalid={error ? true : undefined}
						/>

						{error && <p className='text-destructive text-sm'>{error}</p>}

						<Button type='submit' loading={isSubmitting} fullWidth className='h-10'>
							Continue
						</Button>
					</form>
				</div>

				<p className='text-dim mt-auto text-center text-xs'>
					Need an account? Ask a project owner to invite you, or{' '}
					<Link to='/login' className='underline underline-offset-4'>
						contact support
					</Link>
					.
				</p>
			</div>
		</main>
	)
}

export const Route = createFileRoute('/login')({
	component: Login,
})

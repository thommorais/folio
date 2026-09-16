import type { UserId } from './project'

export type CurrentUser = {
	readonly id: UserId
	readonly email: string
	readonly name: string
	readonly avatarUrl: string | undefined
}

/** `unknown` and `anonymous` are distinct so a restored session shows a spinner, not a login flash. */
export type Session =
	| { readonly status: 'unknown' }
	| { readonly status: 'anonymous' }
	| { readonly status: 'authenticated'; readonly user: CurrentUser }

export const anonymous: Session = { status: 'anonymous' }

export const unknownSession: Session = { status: 'unknown' }

export const authenticated = (user: CurrentUser): Session => ({ status: 'authenticated', user })

export const userOf = (session: Session): CurrentUser | undefined =>
	session.status === 'authenticated' ? session.user : undefined

export const initialsOf = (user: CurrentUser): string => {
	const source = user.name.trim() || user.email
	const parts = source.split(/[\s@._-]+/u).filter(Boolean)
	const initials = parts.slice(0, 2).map(part => part[0] ?? '')
	return initials.join('').toUpperCase()
}

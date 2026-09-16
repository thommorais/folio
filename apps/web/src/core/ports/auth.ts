import type { CurrentUser, Session } from '../domain/session'

export type Credentials = {
	readonly email: string
	readonly password: string
}

export type AuthError =
	| { readonly kind: 'invalid-credentials' }
	| { readonly kind: 'unverified' }
	| { readonly kind: 'network' }
	| { readonly kind: 'unknown'; readonly message: string }

export type SignInResult =
	| { readonly ok: true; readonly user: CurrentUser }
	| { readonly ok: false; readonly error: AuthError }

export type AuthPort = {
	readonly current: () => Session
	readonly signIn: (credentials: Credentials) => Promise<SignInResult>
	readonly signOut: () => void
	readonly subscribe: (listener: (session: Session) => void) => () => void
}

import { useCallback, useSyncExternalStore } from 'react'
import type { Session } from '_/core/domain/session'
import { userOf } from '_/core/domain/session'
import type { Credentials } from '_/core/ports/auth'
import { useContainer } from './container'

export const useSession = (): Session => {
	const { auth } = useContainer()

	return useSyncExternalStore(auth.subscribe, auth.current, auth.current)
}

export const useCurrentUser = () => userOf(useSession())

export const useSignIn = () => {
	const { auth } = useContainer()

	return useCallback((credentials: Credentials) => auth.signIn(credentials), [auth])
}

export const useSignOut = () => {
	const { auth } = useContainer()

	return useCallback(() => auth.signOut(), [auth])
}

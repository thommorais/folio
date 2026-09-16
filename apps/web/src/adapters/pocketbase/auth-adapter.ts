import { ClientResponseError } from 'pocketbase'
import type { CurrentUser, Session } from '_/core/domain/session'
import { anonymous, authenticated } from '_/core/domain/session'
import { userId } from '_/core/domain/project'
import type { AuthError, AuthPort, Credentials, SignInResult } from '_/core/ports/auth'
import { getPocketBaseClient } from './client'

type AuthRecord = {
	id: string
	email?: string
	name?: string
	avatar?: string
	collectionId?: string
}

const toCurrentUser = (record: AuthRecord): CurrentUser => {
	const pb = getPocketBaseClient()

	return {
		id: userId(record.id),
		email: record.email ?? '',
		name: record.name ?? '',
		avatarUrl:
			record.avatar && record.collectionId
				? pb.files.getURL({ id: record.id, collectionId: record.collectionId }, record.avatar)
				: undefined,
	}
}

const toAuthError = (cause: unknown): AuthError => {
	if (cause instanceof ClientResponseError) {
		if (cause.status === 0) return { kind: 'network' }
		if (cause.status === 400) return { kind: 'invalid-credentials' }
		if (cause.status === 403) return { kind: 'unverified' }
		return { kind: 'unknown', message: cause.message }
	}
	return { kind: 'unknown', message: cause instanceof Error ? cause.message : 'Sign in failed' }
}

export const createAuthAdapter = (): AuthPort => {
	// useSyncExternalStore compares by identity: a fresh object per read loops.
	let snapshot: Session = anonymous
	let snapshotKey = ''

	const pb = getPocketBaseClient()

	const read = (): Session => {
		const record = pb.authStore.record
		const key = pb.authStore.isValid && record ? `${record.id}:${pb.authStore.token}` : ''

		if (key !== snapshotKey) {
			snapshotKey = key
			snapshot = key === '' ? anonymous : authenticated(toCurrentUser(record as AuthRecord))
		}

		return snapshot
	}

	return {
		current: read,

		signIn: async ({ email, password }: Credentials): Promise<SignInResult> => {
			try {
				const result = await pb.collection('users').authWithPassword(email, password)
				return { ok: true, user: toCurrentUser(result.record as AuthRecord) }
			} catch (cause) {
				return { ok: false, error: toAuthError(cause) }
			}
		},

		signOut: () => {
			pb.authStore.clear()
		},

		subscribe: listener => pb.authStore.onChange(() => listener(read())),
	}
}

import { useEffect, useState } from 'react'
import type { SharedItem } from '_/core/domain/share'
import type { Result } from '_/lib/result'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

export const SHARE_POLL_MS = 10_000

type SharedState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly item: SharedItem }
	| { readonly status: typeof Status.Gone }
	| { readonly status: typeof Status.Failed; readonly message: string }

const next = (current: SharedState, result: Result<SharedItem | undefined>): SharedState => {
	if (!result.success) {
		return current.status === Status.Ready ? current : { status: Status.Failed, message: result.error.message }
	}

	return result.value === undefined ? { status: Status.Gone } : { status: Status.Ready, item: result.value }
}

export const useShared = (token: string): SharedState => {
	const { shares } = useContainer()
	const [state, setState] = useState<SharedState>({ status: Status.Loading })

	useEffect(() => {
		let sequence = 0
		let applied = 0
		let revoked = false
		setState({ status: Status.Loading })

		const load = async () => {
			if (revoked) return
			sequence += 1
			const attempt = sequence
			const result = await shares.open(token)

			if (attempt <= applied) return
			applied = attempt

			if (result.success && result.value === undefined) revoked = true
			setState(current => next(current, result))
		}

		const poll = () => {
			if (document.visibilityState === 'visible') void load()
		}

		void load()
		const timer = setInterval(poll, SHARE_POLL_MS)
		document.addEventListener('visibilitychange', poll)

		return () => {
			applied = Number.POSITIVE_INFINITY
			clearInterval(timer)
			document.removeEventListener('visibilitychange', poll)
		}
	}, [token, shares])

	return state
}

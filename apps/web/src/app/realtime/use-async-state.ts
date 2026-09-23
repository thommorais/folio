import type { Result } from '_/lib/result'
import { useCallback, useEffect, useEffectEvent, useRef, useState } from 'react'
import { Status } from '_/lib/async-status'

export type AsyncState<T> =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly data: T }
	| { readonly status: typeof Status.Failed; readonly message: string }

type Load<T> = () => Promise<Result<T>>

export const useAsyncState = <T>(load: Load<T>, deps: readonly unknown[], skip = false) => {
	const [state, setState] = useState<AsyncState<T>>({ status: Status.Loading })
	const generation = useRef(0)

	const run = useEffectEvent(async (attempt: number) => {
		const result = await load()

		if (attempt !== generation.current) return

		setState(
			result.success
				? { status: Status.Ready, data: result.value }
				: { status: Status.Failed, message: result.error.message },
		)
	})

	const effectDeps = [...deps, skip]

	const refetch = useCallback(async () => {
		await run(generation.current)
	}, [])

	const patch = useCallback((change: (data: T) => T) => {
		setState(current =>
			current.status === Status.Ready ? { status: Status.Ready, data: change(current.data) } : current,
		)
	}, [])

	// oxlint-disable-next-line react-hooks/exhaustive-deps
	useEffect(() => {
		generation.current += 1
		setState({ status: Status.Loading })
		if (skip) return
		void run(generation.current)

		return () => {
			generation.current += 1
		}
	}, effectDeps)

	return { state, refetch, patch }
}

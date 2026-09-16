import type { ConnectionPort } from '_/core/ports/connection'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { Result } from '_/lib/result'
import { useEffect, useEffectEvent, useRef, useState } from 'react'
import { useAsyncState } from './use-async-state'
import { useSubscription } from './use-subscription'

type Titled = { readonly id: string; readonly title: string }

export type RecordState<T> =
	| { readonly status: 'idle' }
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly data: T }
	| { readonly status: 'gone'; readonly title: string }
	| { readonly status: 'failed'; readonly message: string }

type Options<T extends Titled> = {
	readonly load: () => Promise<Result<T>>
	readonly subscribe: (id: string, onChange: (record: T) => void, onGone: () => void) => Promise<Result<Unsubscribe>>
	readonly connection: ConnectionPort
	readonly deps: readonly unknown[]
	readonly skip?: boolean
}

export const useLiveRecord = <T extends Titled>({
	load,
	subscribe,
	connection,
	deps,
	skip = false,
}: Options<T>): RecordState<T> => {
	const { state, refetch, patch } = useAsyncState(load, deps, skip)
	const [gone, setGone] = useState<string | undefined>(undefined)

	// oxlint-disable-next-line react-hooks/exhaustive-deps
	useEffect(() => {
		setGone(undefined)
	}, deps)

	const latest = useRef<T | undefined>(undefined)
	if (state.status === 'ready') latest.current = state.data

	const open = useEffectEvent(async () => {
		if (skip || state.status !== 'ready') {
			return { success: true, value: async () => {} } as Result<Unsubscribe>
		}

		const result = await subscribe(
			state.data.id,
			record => patch(() => record),
			() => setGone(latest.current?.title ?? ''),
		)

		if (result.success) void refetch()
		return result
	})

	useSubscription(open, [state.status === 'ready' ? state.data.id : undefined])

	useEffect(() => connection.onReconnect(() => void refetch()), [connection, refetch])

	if (skip) return { status: 'idle' }

	if (gone !== undefined) return { status: 'gone', title: gone }

	return state
}

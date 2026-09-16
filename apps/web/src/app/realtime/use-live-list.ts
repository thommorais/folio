import type { ConnectionPort } from '_/core/ports/connection'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { Result } from '_/lib/result'
import type { ActionEvent } from '_/types'
import { useEffect, useEffectEvent } from 'react'
import { type AsyncState, useAsyncState } from './use-async-state'
import { useSubscription } from './use-subscription'

type Entity = Record<'id', unknown>

type Options<T extends Entity> = {
	readonly load: () => Promise<Result<readonly T[]>>
	readonly subscribe: (update: (row: T, action: ActionEvent) => void) => Promise<Result<Unsubscribe>>
	readonly fold: (rows: readonly T[], row: T, action: ActionEvent) => readonly T[]
	readonly connection: ConnectionPort
	readonly deps: readonly unknown[]
}

export const useLiveList = <T extends Entity>({
	load,
	subscribe,
	fold,
	connection,
	deps,
}: Options<T>): AsyncState<readonly T[]> => {
	const { state, refetch, patch } = useAsyncState(load, deps)

	const update = useEffectEvent((row: T, action: ActionEvent) => {
		patch(rows => fold(rows, row, action))
	})

	const open = useEffectEvent(async () => {
		const result = await subscribe(update)
		if (result.success) void refetch()
		return result
	})

	useSubscription(open, deps)

	useEffect(() => connection.onReconnect(() => void refetch()), [connection, refetch])

	return state
}

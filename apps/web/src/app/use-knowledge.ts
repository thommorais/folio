import { useEffect, useEffectEvent } from 'react'
import type { Knowledge } from '_/core/ports/knowledge'
import { Status } from '_/lib/async-status'
import { useContainer } from './container'
import { useAsyncState } from './realtime/use-async-state'
import { useLiveRecord } from './realtime/use-live-record'
import { useSubscription } from './realtime/use-subscription'

type ListState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly notes: readonly Knowledge[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

type NoteState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly note: Knowledge }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useKnowledgeList = (text: string): ListState => {
	const { knowledge, connection } = useContainer()
	const { state, refetch } = useAsyncState(() => knowledge.list({ text }), [knowledge, text])

	const open = useEffectEvent(async () => {
		const result = await knowledge.subscribeToList(() => void refetch())
		if (result.success) void refetch()
		return result
	})

	useSubscription(open, [knowledge])

	useEffect(() => connection.onReconnect(() => void refetch()), [connection, refetch])

	return state.status === Status.Ready ? { status: Status.Ready, notes: state.data } : state
}

export const useKnowledge = (ref: string): NoteState => {
	const { knowledge, connection } = useContainer()

	const state = useLiveRecord<Knowledge>({
		load: () => knowledge.get(ref),
		subscribe: (id, onChange, onGone) => knowledge.subscribeToRecord(id, onChange, onGone),
		connection,
		deps: [knowledge, ref],
	})

	return state.status === Status.Ready ? { status: Status.Ready, note: state.data } : state
}

import type { Unsubscribe } from '_/core/ports/subscription'
import type { Result } from '_/lib/result'
import { useEffect, useEffectEvent, useState } from 'react'
import { useSubscription } from './realtime/use-subscription'
import { useContainer } from './container'
import { collectCounts, type CountsState } from './counts'

export const useCounts = (project: string): CountsState => {
	const { entries, issues, plans, connection } = useContainer()
	const [state, setState] = useState<CountsState>({ status: 'loading' })

	const load = useEffectEvent(async () => {
		const results = await Promise.all([
			issues.count(project, { kind: 'ticket' }),
			plans.count(project),
			issues.count(project, { kind: 'todo' }),
			entries.count(project, { kind: 'journal' }),
			entries.count(project, { kind: 'doc' }),
		])

		setState(collectCounts(results))
	})

	useEffect(() => {
		setState({ status: 'loading' })
		void load()
	}, [project])

	// The tiles sit beside lists that update themselves, so a count that only
	// loaded once would drift out of step with the rows right next to it. The
	// event carries one record, not a total, so recount instead of adjusting.
	const recount = useEffectEvent(() => {
		void load()
	})

	const open = useEffectEvent(async (): Promise<Result<Unsubscribe>> => {
		const opened = await Promise.all([
			issues.subscribeToList(project, recount),
			plans.subscribeToList(project, recount),
			entries.subscribeToList(project, recount),
		])

		const closers = opened.filter(result => result.success).map(result => result.value)

		return {
			success: true,
			value: async () => {
				await Promise.all(closers.map(close => close()))
			},
		}
	})

	useSubscription(open, [project, issues, plans, entries])

	useEffect(() => connection.onReconnect(() => void load()), [connection])

	return state
}

import { ADDRESSABLE_KINDS, KIND } from '_/core/domain/entry';
import type { Unsubscribe } from '_/core/ports/subscription';
import { Status } from '_/lib/async-status';
import type { Result } from '_/lib/result';
import { useEffect, useEffectEvent, useState } from 'react';
import { useContainer } from './container';
import { collectCounts, type CountsState } from './counts';
import { useSubscription } from './realtime/use-subscription';

export const useCounts = (project: string): CountsState => {
	const { entries, issues, plans, connection } = useContainer()
	const [state, setState] = useState<CountsState>({ status: Status.Loading })

	const load = useEffectEvent(async () => {
		const results = await Promise.all([
			issues.count(project, { kind: KIND.TICKET }),
			plans.count(project),
			issues.count(project, { kind: KIND.TODO }),
			// The journal lists both kinds, so its tile counts both.
			entries.count(project, { kinds: ADDRESSABLE_KINDS }),
		])

		setState(collectCounts(results))
	})

	useEffect(() => {
		setState({ status: Status.Loading })
		void load()
	}, [project])

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

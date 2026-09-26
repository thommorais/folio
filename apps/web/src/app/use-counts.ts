import { ADDRESSABLE_KINDS } from '_/core/domain/entry';
import type { Unsubscribe } from '_/core/ports/subscription';
import { Status } from '_/lib/async-status';
import type { Result } from '_/lib/result';
import { useEffect, useEffectEvent, useState } from 'react';
import { useContainer } from './container';
import { collectCounts, type CountsState } from './counts';
import { useSubscription } from './realtime/use-subscription';
import { DEFAULT_ISSUE_STATUSES, ISSUE_KIND } from '_/core/domain/issue';
import { DEFAULT_PLAN_STATUSES } from '_/core/domain/plan';

export const useCounts = (project: string): CountsState => {
	const { entries, issues, plans, connection } = useContainer()
	const [state, setState] = useState<CountsState>({ status: Status.Loading })

	const load = useEffectEvent(async () => {
		const results = await Promise.all([
			issues.count(project, { kind: ISSUE_KIND.TICKET, status: DEFAULT_ISSUE_STATUSES }),
			plans.count(project, { status: DEFAULT_PLAN_STATUSES }),
			issues.count(project, { kind: ISSUE_KIND.TODO, status: DEFAULT_ISSUE_STATUSES }),
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

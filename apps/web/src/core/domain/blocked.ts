import type { Issue } from './issue'
import { isTerminal } from './issue'

// blocked is derived from the blockers present in the set, so a live update to
// one issue changes it for every issue that depends on it. The event only
// carries the one that changed, which leaves the rest stale until they are
// recomputed here.
//
// A blocker outside the set cannot be judged and so does not block, the same
// rule the adapter applies when it first loads a list.
const statusOf = (issues: readonly Issue[]): ReadonlyMap<Issue['id'], Issue['status']> =>
	new Map(issues.map(issue => [issue.id, issue.status]))

const openAmong = (issue: Issue, status: ReadonlyMap<Issue['id'], Issue['status']>): number =>
	issue.dependsOn.filter(blocker => {
		const state = status.get(blocker)
		return state !== undefined && !isTerminal(state)
	}).length

export const openBlockers = (issue: Issue, known: readonly Issue[]): number => openAmong(issue, statusOf(known))

export const withBlocked = (issues: readonly Issue[]): readonly Issue[] => {
	const status = statusOf(issues)

	return issues.map(issue => {
		const blocked = openAmong(issue, status) > 0

		// Returned unchanged when the flag already agrees, so a render is only
		// triggered for the rows that actually moved.
		return blocked === issue.blocked ? issue : { ...issue, blocked }
	})
}

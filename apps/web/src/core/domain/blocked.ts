import type { Issue } from './issue'
import { isTerminal } from './issue'

// blocked is derived from the blockers present in the set, so a live update to
// one issue changes it for every issue that depends on it. The event only
// carries the one that changed, which leaves the rest stale until they are
// recomputed here.
//
// A blocker outside the set cannot be judged and so does not block, the same
// rule the adapter applies when it first loads a list.
export const withBlocked = (issues: readonly Issue[]): readonly Issue[] => {
	const status = new Map(issues.map(issue => [issue.id, issue.status]))

	return issues.map(issue => {
		const blocked = issue.dependsOn.some(blocker => {
			const state = status.get(blocker)
			return state !== undefined && !isTerminal(state)
		})

		// Returned unchanged when the flag already agrees, so a render is only
		// triggered for the rows that actually moved.
		return blocked === issue.blocked ? issue : { ...issue, blocked }
	})
}

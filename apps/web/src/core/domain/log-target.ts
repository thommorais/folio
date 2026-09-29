import type { Entry } from './entry'
import { ISSUE_KIND, type Issue } from './issue'
import type { Plan } from './plan'

export type LogTarget = {
	readonly kind: 'ticket' | 'todo' | 'plan'
	readonly ref: string
	readonly title: string
}

export const logTarget = (entry: Entry, issues: readonly Issue[], plans: readonly Plan[]): LogTarget | undefined => {
	if (entry.issueId !== undefined) {
		const found = issues.find(candidate => candidate.id === entry.issueId)

		return found === undefined
			? undefined
			: { kind: found.kind === ISSUE_KIND.TODO ? 'todo' : 'ticket', ref: found.slug, title: found.title }
	}

	if (entry.planId !== undefined) {
		const found = plans.find(candidate => candidate.id === entry.planId)

		return found === undefined ? undefined : { kind: 'plan', ref: found.id, title: found.title }
	}

	return undefined
}

import { ISSUE_KIND, type Issue } from './issue'

export const isTopLevel = (issue: Issue): boolean =>
	issue.parentId === undefined && (issue.kind === ISSUE_KIND.TICKET || issue.planId === undefined)

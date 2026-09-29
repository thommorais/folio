import { describe, expect, it } from 'vitest'
import { ISSUE_KIND, type Issue, type IssueId } from './issue'
import type { PlanId } from './plan'
import { isTopLevel } from './top-level'

const issue = (over: Partial<Issue> = {}): Issue =>
	({
		id: 'a' as IssueId,
		kind: ISSUE_KIND.TICKET,
		planId: undefined,
		parentId: undefined,
		...over,
	}) as Issue

describe('isTopLevel', () => {
	it('keeps a ticket with no parent', () => {
		expect(isTopLevel(issue())).toBe(true)
	})

	it('drops a ticket under another ticket', () => {
		expect(isTopLevel(issue({ parentId: 'p' as IssueId }))).toBe(false)
	})

	it('keeps a ticket that belongs to a plan but has no parent', () => {
		expect(isTopLevel(issue({ planId: 'pl' as PlanId }))).toBe(true)
	})

	it('keeps a todo with no plan and no parent', () => {
		expect(isTopLevel(issue({ kind: ISSUE_KIND.TODO }))).toBe(true)
	})

	it('drops a todo under a ticket', () => {
		expect(isTopLevel(issue({ kind: ISSUE_KIND.TODO, parentId: 'p' as IssueId }))).toBe(false)
	})

	it('drops a todo under a plan', () => {
		expect(isTopLevel(issue({ kind: ISSUE_KIND.TODO, planId: 'pl' as PlanId }))).toBe(false)
	})
})

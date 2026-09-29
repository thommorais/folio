import { describe, expect, it } from 'vitest'
import type { Entry } from './entry'
import { ISSUE_KIND, type Issue, type IssueId } from './issue'
import type { Plan, PlanId } from './plan'
import { logTarget } from './log-target'

const entry = (over: Partial<Entry> = {}): Entry => ({ issueId: undefined, planId: undefined, ...over }) as Entry

const issue = (id: string, kind: Issue['kind']): Issue => ({ id: id as IssueId, kind, slug: `${id}-slug`, title: `${id} title` }) as Issue

const plan = (id: string): Plan => ({ id: id as PlanId, title: `${id} title` }) as Plan

describe('logTarget', () => {
	it('points a log written against a ticket at the ticket', () => {
		const got = logTarget(entry({ issueId: 'k1' as IssueId }), [issue('k1', ISSUE_KIND.TICKET)], [])

		expect(got).toEqual({ kind: 'ticket', ref: 'k1-slug', title: 'k1 title' })
	})

	it('points a log written against a todo at the todo', () => {
		const got = logTarget(entry({ issueId: 't1' as IssueId }), [issue('t1', ISSUE_KIND.TODO)], [])

		expect(got).toEqual({ kind: 'todo', ref: 't1-slug', title: 't1 title' })
	})

	it('points a log written against a plan at the plan by id', () => {
		const got = logTarget(entry({ planId: 'p1' as PlanId }), [], [plan('p1')])

		expect(got).toEqual({ kind: 'plan', ref: 'p1', title: 'p1 title' })
	})

	it('has no target when the log belongs to nothing', () => {
		expect(logTarget(entry(), [issue('k1', ISSUE_KIND.TICKET)], [plan('p1')])).toBeUndefined()
	})

	it('has no target when the record it names cannot be found', () => {
		expect(logTarget(entry({ issueId: 'gone' as IssueId }), [], [])).toBeUndefined()
	})
})

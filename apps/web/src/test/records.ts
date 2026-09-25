import type { Cycle, CycleId } from '_/core/domain/cycle'
import type { Entry, EntryId } from '_/core/domain/entry'
import type { Issue, IssueId } from '_/core/domain/issue'
import type { ProjectId } from '_/core/domain/project'

const at = new Date('2026-01-01')

export const anIssue = (id: string, over: Partial<Issue> = {}): Issue => ({
	id: id as IssueId,
	kind: 'ticket',
	projectId: 'p1' as ProjectId,
	planId: undefined,
	slug: id,
	title: id,
	body: '',
	status: 'open',
	priority: 'medium',
	size: undefined,
	assignee: undefined,
	tags: [],
	position: 0,
	dueDate: undefined,
	wayfinder: undefined,
	externalRef: '',
	resolution: '',
	resolutionEntry: undefined,
	createdBy: undefined,
	createdAt: at,
	updatedAt: at,
	parentId: undefined,
	dependsOn: [],
	relatedTo: [],
	blocked: false,
	...over,
})

export const anEntry = (id: string, over: Partial<Entry> = {}): Entry => ({
	id: id as EntryId,
	kind: 'journal',
	projectId: 'p1' as ProjectId,
	issueId: undefined,
	planId: undefined,
	cycleId: undefined,
	slug: '',
	title: '',
	body: '',
	branch: '',
	pr: '',
	externalRef: '',
	tags: [],
	createdBy: undefined,
	createdAt: at,
	updatedAt: at,
	...over,
})

export const aCycle = (id: string, over: Partial<Cycle> = {}): Cycle => ({
	id: id as CycleId,
	projectId: 'p1' as ProjectId,
	ticketId: 'work' as IssueId,
	ordinal: 1,
	phase: 'plan',
	resolution: '',
	mapId: undefined,
	createdBy: undefined,
	createdAt: at,
	updatedAt: at,
	closedAt: undefined,
	...over,
})

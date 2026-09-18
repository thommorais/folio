import { sortExpr } from './sort'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import { planId as toPlanId } from '_/core/domain/plan'
import type { Issue, IssueKind, IssueStatus, Priority, Size, WayfinderType } from '_/core/domain/issue'
import { isSize, issueId as toIssueId } from '_/core/domain/issue'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { IssueFilter, IssuesPort } from '_/core/ports/issues'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournIssuesResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { keyed } from './request-key'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { countRows } from './count-rows'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'
import { linksOf, tagsByTarget } from './relations'

type IssueRecord = JournIssuesResponse<string[]>

type IssueColumns = {
	'project.slug': string
	kind: IssueKind
	plan: string
	slug: string
	title: string
	body: string
	status: IssueStatus
	priority: Priority
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toIssue = (record: IssueRecord): Issue => ({
	id: toIssueId(record.id),
	kind: record.kind as IssueKind,
	projectId: toProjectId(record.project),
	planId: record.plan ? toPlanId(record.plan) : undefined,
	slug: record.slug,
	title: record.title,
	body: record.body ?? '',
	status: record.status as IssueStatus,
	priority: record.priority as Priority,
	size: record.size && isSize(record.size) ? (record.size as Size) : undefined,
	assignee: record.assignee ? toUserId(record.assignee) : undefined,
	tags: [],
	position: record.position ?? 0,
	dueDate: record.due_date ? new Date(record.due_date) : undefined,
	wayfinder: record.wayfinder ? (record.wayfinder as WayfinderType) : undefined,
	externalRef: record.external_ref ?? '',
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
	parentId: undefined,
	dependsOn: [],
	relatedTo: [],
	blocked: false,
})

const hydrate = async (issues: readonly Issue[]): Promise<readonly Issue[]> => {
	if (issues.length === 0) return issues

	const ids = issues.map(issue => issue.id as string)
	const [tags, links] = await Promise.all([tagsByTarget('issue', ids), linksOf(ids)])

	const status = new Map(issues.map(issue => [issue.id as string, issue.status]))

	return issues.map(issue => {
		const id = issue.id as string
		const dependsOn = links.dependsOn.get(id) ?? []
		const parent = links.parentOf.get(id)

		return {
			...issue,
			tags: tags.get(id) ?? [],
			parentId: parent ? toIssueId(parent) : undefined,
			dependsOn: dependsOn.map(toIssueId),
			relatedTo: (links.relatedTo.get(id) ?? []).map(toIssueId),
			// A blocker outside the loaded set cannot be judged, so it does not block.
			blocked: dependsOn.some(blocker => {
				const state = status.get(blocker)
				return state !== undefined && state !== 'done' && state !== 'cancelled'
			}),
		}
	})
}

const hydrateOne = async (issue: Issue): Promise<Issue> => {
	const [hydrated] = await hydrate([issue])
	return hydrated ?? issue
}

const columns = (project: string, filter: IssueFilter) =>
	filterFor<IssueColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'kind', comparator: 'eq', value: filter.kind },
		{ field: 'plan', comparator: 'eq', value: filter.planId },
		{ field: 'status', comparator: 'anyOf', value: filter.status },
		{ field: 'priority', comparator: 'eq', value: filter.priority },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

const matchesTags = (issue: Issue, tags: readonly string[] | undefined): boolean => {
	if (!tags || tags.length === 0) return true
	const have = new Set(issue.tags.map(tag => tag.toLowerCase()))
	return tags.every(tag => have.has(tag.toLowerCase()))
}

const childrenOf = (issues: readonly Issue[], parentId: string | undefined): readonly Issue[] =>
	parentId === undefined ? issues : issues.filter(issue => issue.parentId === parentId)

export const createIssuesAdapter = (): IssuesPort => {
	const client = getPocketBaseClient()
	const collection = () => client.collection(Collections.JournIssues)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				countRows(collection(), keyed('issues.count', { filter: client.filter(expr, params) })),
			)

			return error ? err(new Error(`Failed to count issues: ${error.message}`, { cause: error })) : ok(data)
		},

		list: async (project, filter = {}): Promise<Result<readonly Issue[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<IssueRecord>(
					collection(),
					filter,
					keyed(
						'issues.list',
						{
							filter: client.filter(expr, params),
							sort: sortExpr(filter.sort, '-created'),
						},
						// parentId and tags narrow the rows after they arrive, and the
						// paging is applied by paginate, so none of it reaches the
						// filter expression the key is otherwise built from.
						[filter.limit, filter.offset, filter.parentId, filter.tags],
					),
				),
			)
			if (error) return err(new Error(`Failed to list issues: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(hydrate(data.map(toIssue)))
			if (hydrated.error) {
				return err(new Error(`Failed to load issue relations: ${hydrated.error.message}`, { cause: hydrated.error }))
			}

			return ok(childrenOf(hydrated.data, filter.parentId).filter(issue => matchesTags(issue, filter.tags)))
		},

		get: async (project, slug): Promise<Result<Issue>> => {
			const { expr, params } = filterFor<IssueColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'slug', comparator: 'eq', value: slug },
			])

			const { data, error } = await tryCatch(
				collection().getFirstListItem<IssueRecord>(client.filter(expr, params), keyed('issues.get', {})),
			)
			if (error) return err(new Error(`Failed to load issue ${slug}: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(hydrateOne(toIssue(data)))
			return hydrated.error
				? err(new Error(`Failed to load issue relations: ${hydrated.error.message}`, { cause: hydrated.error }))
				: ok(hydrated.data)
		},

		getById: async (project, id): Promise<Result<Issue>> => {
			const { expr, params } = filterFor<IssueColumns>()([{ field: 'project.slug', comparator: 'eq', value: project }])

			const { data, error } = await tryCatch(
				collection().getOne<IssueRecord>(id, keyed(`issues.getById.${id}`, { filter: client.filter(expr, params) })),
			)
			if (error) return err(new Error(`Failed to load issue ${id}: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(hydrateOne(toIssue(data)))
			return hydrated.error
				? err(new Error(`Failed to load issue relations: ${hydrated.error.message}`, { cause: hydrated.error }))
				: ok(hydrated.data)
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection().subscribe<IssueRecord>(
					'*',
					event => {
						void hydrateOne(toIssue(event.record)).then(issue => {
							update(issue, event.action as ActionEvent)
						})
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to issues: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(
				collection(),
				id,
				toIssue,
				issue => {
					void hydrateOne(issue).then(onChange)
				},
				onGone,
				'issue',
			),
	}
}

import { sortExpr } from './sort'
import { projectId as toProjectId, userId as toUserId } from '_/core/domain/project'
import { planId as toPlanId } from '_/core/domain/plan'
import { cycleId as toCycleId } from '_/core/domain/cycle'
import { issueId as toIssueId } from '_/core/domain/issue'
import type { Entry, EntryKind } from '_/core/domain/entry'
import { entryId as toEntryId } from '_/core/domain/entry'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { EntriesPort, EntryFilter } from '_/core/ports/entries'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournEntriesResponse } from '_/pocketbase-types'
import type { ActionEvent } from '_/types'
import { getPocketBaseClient } from './client'
import { subscribeToRecord as subscribe } from './subscribe-to-record'
import { countRows } from './count-rows'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'
import { tagsByTarget } from './relations'

type EntryRecord = JournEntriesResponse<unknown, string[]>

type EntryColumns = {
	'project.slug': string
	kind: EntryKind
	issue: string
	plan: string
	cycle: string
	branch: string
	external_ref: string
	slug: string
	title: string
	body: string
	created: Date
	updated: Date
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toEntry = (record: EntryRecord): Entry => ({
	id: toEntryId(record.id),
	kind: record.kind as EntryKind,
	projectId: toProjectId(record.project),
	issueId: record.issue ? toIssueId(record.issue) : undefined,
	planId: record.plan ? toPlanId(record.plan) : undefined,
	cycleId: record.cycle ? toCycleId(record.cycle) : undefined,
	slug: record.slug ?? '',
	title: record.title ?? '',
	body: record.body ?? '',
	branch: record.branch ?? '',
	pr: record.pr ?? '',
	externalRef: record.external_ref ?? '',
	tags: [],
	createdBy: record.created_by ? toUserId(record.created_by) : undefined,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

const withTags = async (entries: readonly Entry[]): Promise<readonly Entry[]> => {
	if (entries.length === 0) return entries

	const tags = await tagsByTarget(
		'entry',
		entries.map(entry => entry.id as string),
	)

	return entries.map(entry => ({ ...entry, tags: tags.get(entry.id as string) ?? [] }))
}

const withTagsOne = async (entry: Entry): Promise<Entry> => {
	const [hydrated] = await withTags([entry])
	return hydrated ?? entry
}

const columns = (project: string, filter: EntryFilter) =>
	filterFor<EntryColumns>()([
		{ field: 'project.slug', comparator: 'eq', value: project },
		{ field: 'kind', comparator: 'eq', value: filter.kind },
		{ field: 'issue', comparator: 'eq', value: filter.issueId },
		{ field: 'plan', comparator: 'eq', value: filter.planId },
		{ field: 'cycle', comparator: 'eq', value: filter.cycleId },
		{ field: 'branch', comparator: 'eq', value: filter.branch },
		{ field: 'external_ref', comparator: 'eq', value: filter.externalRef },
		{ field: 'title', comparator: 'contains', value: filter.search },
	])

const matchesTags = (entry: Entry, tags: readonly string[] | undefined): boolean => {
	if (!tags || tags.length === 0) return true
	const have = new Set(entry.tags.map(tag => tag.toLowerCase()))
	return tags.every(tag => have.has(tag.toLowerCase()))
}

export const createEntriesAdapter = (): EntriesPort => {
	const client = getPocketBaseClient()
	const collection = () => client.collection(Collections.JournEntries)

	return {
		count: async (project, filter = {}): Promise<Result<number>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(countRows(collection(), { filter: client.filter(expr, params) }))

			return error ? err(new Error(`Failed to count entries: ${error.message}`, { cause: error })) : ok(data)
		},

		list: async (project, filter = {}): Promise<Result<readonly Entry[]>> => {
			const { expr, params } = columns(project, filter)

			const { data, error } = await tryCatch(
				paginate<EntryRecord>(collection(), filter, {
					filter: client.filter(expr, params),
					sort: sortExpr(filter.sort, '-created'),
				}),
			)
			if (error) return err(new Error(`Failed to list entries: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(withTags(data.map(toEntry)))
			if (hydrated.error) {
				return err(new Error(`Failed to load entry tags: ${hydrated.error.message}`, { cause: hydrated.error }))
			}

			return ok(hydrated.data.filter(entry => matchesTags(entry, filter.tags)))
		},

		get: async (project, slug): Promise<Result<Entry>> => {
			const { expr, params } = filterFor<EntryColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
				{ field: 'slug', comparator: 'eq', value: slug },
			])

			const { data, error } = await tryCatch(collection().getFirstListItem<EntryRecord>(client.filter(expr, params)))
			if (error) return err(new Error(`Failed to load entry ${slug}: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(withTagsOne(toEntry(data)))
			return hydrated.error
				? err(new Error(`Failed to load entry tags: ${hydrated.error.message}`, { cause: hydrated.error }))
				: ok(hydrated.data)
		},

		getById: async (project, id): Promise<Result<Entry>> => {
			const { expr, params } = filterFor<EntryColumns>()([
				{ field: 'project.slug', comparator: 'eq', value: project },
			])

			const { data, error } = await tryCatch(
				collection().getOne<EntryRecord>(id, { filter: client.filter(expr, params) }),
			)
			if (error) return err(new Error(`Failed to load entry ${id}: ${error.message}`, { cause: error }))

			const hydrated = await tryCatch(withTagsOne(toEntry(data)))
			return hydrated.error
				? err(new Error(`Failed to load entry tags: ${hydrated.error.message}`, { cause: hydrated.error }))
				: ok(hydrated.data)
		},

		subscribeToList: async (project, update, filter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(project, filter)

			try {
				const unsubscribe = await collection().subscribe<EntryRecord>(
					'*',
					event => {
						void withTagsOne(toEntry(event.record)).then(entry => {
							update(entry, event.action as ActionEvent)
						})
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to entries: ${message(error)}`))
			}
		},

		subscribeToRecord: async (_project, id, onChange, onGone): Promise<Result<Unsubscribe>> =>
			subscribe(
				collection(),
				id,
				toEntry,
				entry => {
					void withTagsOne(entry).then(onChange)
				},
				onGone,
				'entry',
			),
	}
}

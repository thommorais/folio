import { domainId as toDomainId } from '_/core/domain/domain'
import type { Project } from '_/core/domain/project'
import { projectId as toProjectId } from '_/core/domain/project'
import type { ProjectFilter, ProjectsPort } from '_/core/ports/projects'
import type { Unsubscribe } from '_/core/ports/subscription'
import type { ActionEvent } from '_/types'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { Collections, type JournProjectsResponse } from '_/pocketbase-types'
import { getPocketBaseClient } from './client'
import { filterFor } from './filter-builder'
import { paginate } from './paginate'

type ProjectRecord = JournProjectsResponse

type ProjectColumns = {
	domain: string
	slug: string
	name: string
	descr: string
	archived: boolean
	created: Date
	updated: Date
}

const columns = (filter: ProjectFilter) =>
	filterFor<ProjectColumns>()([
		{ field: 'domain', comparator: 'eq', value: filter.domainId },
		{ field: 'archived', comparator: 'eq', value: filter.includeArchived ? undefined : false },
		{ field: 'name', comparator: 'contains', value: filter.search },
	])

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toProject = (record: ProjectRecord): Project => ({
	id: toProjectId(record.id),
	domainId: toDomainId(record.domain ?? ''),
	slug: record.slug,
	name: record.name,
	descr: record.descr ?? '',
	archived: record.archived ?? false,
	createdAt: new Date(record.created),
	updatedAt: new Date(record.updated),
})

export const createProjectsAdapter = (): ProjectsPort => {
	const client = getPocketBaseClient()
	const projects = () => client.collection(Collections.JournProjects)

	return {
		list: async (filter: ProjectFilter = {}): Promise<Result<readonly Project[]>> => {
			const { expr, params } = columns(filter)

			const { data, error } = await tryCatch(
				paginate<ProjectRecord>(projects(), filter, {
					filter: client.filter(expr, params),
					sort: 'name',
				}),
			)

			return error
				? err(new Error(`Failed to list projects: ${error.message}`, { cause: error }))
				: ok(data.map(toProject))
		},

		get: async (ref: string): Promise<Result<Project>> => {
			const { expr, params } = filterFor<ProjectColumns>()([{ field: 'slug', comparator: 'eq', value: ref }])

			const { data, error } = await tryCatch(projects().getFirstListItem<ProjectRecord>(client.filter(expr, params)))

			return error
				? err(new Error(`Failed to load project ${ref}: ${error.message}`, { cause: error }))
				: ok(toProject(data))
		},

		subscribeToList: async (update, filter: ProjectFilter = {}): Promise<Result<Unsubscribe>> => {
			const { expr, params } = columns(filter)

			try {
				const unsubscribe = await projects().subscribe<ProjectRecord>(
					'*',
					event => {
						update(toProject(event.record), event.action as ActionEvent)
					},
					{ filter: client.filter(expr, params) },
				)

				return ok(unsubscribe)
			} catch (error) {
				return err(new Error(`Failed to subscribe to projects: ${message(error)}`))
			}
		},
	}
}

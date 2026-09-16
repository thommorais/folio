import type { Project } from '_/core/domain/project'

export type Insight = {
	readonly key: string
	readonly before: string
	readonly link: string
	readonly after: string
	readonly to?: string
	readonly params?: Record<string, string>
}

const plural = (count: number, one: string, many: string) => `${count} ${count === 1 ? one : many}`

// Every line is built from the projects list the page already loaded. Nothing
// here reads a count we have not fetched.
export const buildInsights = (projects: readonly Project[]): readonly Insight[] => {
	const active = projects.filter(project => !project.archived)
	const archived = projects.length - active.length
	const insights: Insight[] = []

	if (active.length > 0) {
		insights.push({
			key: 'active',
			before: 'You are a member of ',
			link: plural(active.length, 'active project', 'active projects'),
			after: '.',
		})
	}

	const recent = [...active].sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime()).at(0)

	if (recent !== undefined) {
		insights.push({
			key: 'recent',
			before: 'Last touched: ',
			link: recent.name,
			after: '.',
			to: '/$slug',
			params: { slug: recent.slug },
		})
	}

	const shared = active.filter(project => project.members.length > 1)

	if (shared.length > 0) {
		insights.push({
			key: 'shared',
			before: '',
			link: plural(shared.length, 'project', 'projects'),
			after: ` you share with someone else.`,
		})
	}

	if (archived > 0) {
		insights.push({
			key: 'archived',
			before: 'You have ',
			link: plural(archived, 'archived project', 'archived projects'),
			after: '.',
		})
	}

	if (insights.length === 0) {
		insights.push({
			key: 'empty',
			before: '',
			link: '',
			after: 'No projects yet. Create one to get started.',
		})
	}

	return insights
}

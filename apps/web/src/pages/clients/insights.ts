import type { Client } from '_/core/domain/client'
import type { Domain } from '_/core/domain/domain'

export type Insight = {
	readonly key: string
	readonly before: string
	readonly link: string
	readonly after: string
	readonly to?: string
	readonly params?: Record<string, string>
}

const plural = (count: number, one: string, many: string) => `${count} ${count === 1 ? one : many}`

// Every line is built from the lists the page already loaded. Nothing here
// reads a count we have not fetched.
export const buildInsights = (clients: readonly Client[], domains: readonly Domain[]): readonly Insight[] => {
	const insights: Insight[] = []

	if (clients.length > 0) {
		insights.push({
			key: 'clients',
			before: 'You work across ',
			link: plural(clients.length, 'client', 'clients'),
			after: '.',
		})
	}

	if (domains.length > 0) {
		insights.push({
			key: 'domains',
			before: 'Spanning ',
			link: plural(domains.length, 'domain', 'domains'),
			after: '.',
		})
	}

	const recent = [...domains].sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime()).at(0)
	const recentClient = recent && clients.find(client => client.id === recent.clientId)

	if (recent !== undefined && recentClient !== undefined) {
		insights.push({
			key: 'recent',
			before: 'Last touched: ',
			link: recent.name,
			after: '.',
			to: '/$client/$domain',
			params: { client: recentClient.slug, domain: recent.slug },
		})
	}

	const shared = domains.filter(domain => domain.members.length > 1)

	if (shared.length > 0) {
		insights.push({
			key: 'shared',
			before: '',
			link: plural(shared.length, 'domain', 'domains'),
			after: ` you share with someone else.`,
		})
	}

	if (insights.length === 0) {
		insights.push({
			key: 'empty',
			before: '',
			link: '',
			after: 'No clients yet. Create one to get started.',
		})
	}

	return insights
}

import { Collections } from '_/pocketbase-types'
import type { JournEntryTagsResponse, JournIssueLinksResponse, JournIssueTagsResponse, JournTagsResponse } from '_/pocketbase-types'
import { getPocketBaseClient } from './client'

type TagLink = { readonly tag: string; readonly expand?: { readonly tag?: JournTagsResponse } }

const named = (rows: readonly TagLink[]): readonly string[] => {
	const names: string[] = []
	for (const row of rows) {
		const name = row.expand?.tag?.name
		if (name) names.push(name)
	}
	return names
}

const chunked = (ids: readonly string[]): readonly string[][] => {
	const out: string[][] = []
	for (let i = 0; i < ids.length; i += 40) out.push(ids.slice(i, i + 40))
	return out
}

const orFilter = (field: string, ids: readonly string[]): string =>
	ids.map(id => `${field} = "${id.replaceAll('"', '')}"`).join(' || ')

export const tagsByTarget = async (
	target: 'issue' | 'entry',
	ids: readonly string[],
): Promise<ReadonlyMap<string, readonly string[]>> => {
	const out = new Map<string, readonly string[]>()
	if (ids.length === 0) return out

	const client = getPocketBaseClient()
	const collection = target === 'issue' ? Collections.JournIssueTags : Collections.JournEntryTags

	for (const batch of chunked(ids)) {
		const rows = await client
			.collection(collection)
			.getFullList<(JournIssueTagsResponse | JournEntryTagsResponse) & TagLink>({
				filter: orFilter(target, batch),
				expand: 'tag',
			})

		const grouped = new Map<string, TagLink[]>()
		for (const row of rows) {
			const key = target === 'issue' ? (row as { issue: string }).issue : (row as { entry: string }).entry
			const bucket = grouped.get(key)
			if (bucket) bucket.push(row)
			else grouped.set(key, [row])
		}
		for (const [key, group] of grouped) out.set(key, named(group))
	}

	return out
}

export type Links = {
	readonly parentOf: ReadonlyMap<string, string>
	readonly dependsOn: ReadonlyMap<string, readonly string[]>
	readonly relatedTo: ReadonlyMap<string, readonly string[]>
}

export const linksOf = async (ids: readonly string[]): Promise<Links> => {
	const parentOf = new Map<string, string>()
	const dependsOn = new Map<string, string[]>()
	const relatedTo = new Map<string, string[]>()
	if (ids.length === 0) return { parentOf, dependsOn, relatedTo }

	const client = getPocketBaseClient()

	for (const batch of chunked(ids)) {
		const rows = await client.collection(Collections.JournIssueLinks).getFullList<JournIssueLinksResponse>({
			filter: `(${orFilter('from', batch)}) || (${orFilter('to', batch)})`,
		})

		for (const row of rows) {
			if (row.kind === 'parent') {
				parentOf.set(row.from, row.to)
				continue
			}
			if (row.kind === 'blocks') {
				const bucket = dependsOn.get(row.to) ?? []
				bucket.push(row.from)
				dependsOn.set(row.to, bucket)
				continue
			}
			for (const [key, other] of [
				[row.from, row.to] as const,
				[row.to, row.from] as const,
			]) {
				const bucket = relatedTo.get(key) ?? []
				bucket.push(other)
				relatedTo.set(key, bucket)
			}
		}
	}

	return { parentOf, dependsOn, relatedTo }
}

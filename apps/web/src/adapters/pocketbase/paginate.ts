export type PageRequest = {
	readonly limit?: number
	readonly offset?: number
}

type ListOptions = Record<string, unknown>

type Listable<T> = {
	getFullList: (options?: ListOptions) => Promise<T[]>
	getList: (page: number, perPage: number, options?: ListOptions) => Promise<{ items: T[] }>
}

// PocketBase pages by number, the ports by offset, so a non-multiple offset
// would land mid-page. Exact pages hit the API; the rest fetch and slice.
export const paginate = async <T>(collection: Listable<T>, page: PageRequest, options: ListOptions): Promise<T[]> => {
	const { limit, offset = 0 } = page

	if (limit === undefined) {
		const rows = await collection.getFullList(options)
		return offset > 0 ? rows.slice(offset) : rows
	}

	if (offset % limit === 0) {
		const { items } = await collection.getList(offset / limit + 1, limit, options)
		return items
	}

	const rows = await collection.getFullList(options)
	return rows.slice(offset, offset + limit)
}

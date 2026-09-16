type ListOptions = Record<string, unknown>

type Countable = {
	getList: (page: number, perPage: number, options?: ListOptions) => Promise<{ totalItems: number }>
}

// getFullList sets skipTotal, so the list path can never hand back a total and
// counting through it would mean paying for every row. One row is the smallest
// page the API accepts, and totalItems comes back regardless of how few we ask
// for, so the response size stays flat however large the project gets.
export const countRows = async (collection: Countable, options: ListOptions): Promise<number> => {
	const { totalItems } = await collection.getList(1, 1, options)
	return totalItems
}

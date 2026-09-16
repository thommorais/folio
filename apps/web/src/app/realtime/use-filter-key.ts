const stable = (value: unknown): unknown => {
	if (Array.isArray(value)) return value.map(stable)

	if (value === null || typeof value !== 'object') return value

	return Object.fromEntries(
		Object.entries(value as Record<string, unknown>)
			.filter(([, entry]) => entry !== undefined)
			.sort(([left], [right]) => (left < right ? -1 : 1))
			.map(([key, entry]) => [key, stable(entry)]),
	)
}

export const filterKey = (filter: unknown): string => JSON.stringify(stable(filter ?? {}))

export const useFilterKey = (filter: unknown): string => filterKey(filter)

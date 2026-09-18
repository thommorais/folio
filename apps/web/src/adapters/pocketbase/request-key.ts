type ListOptions = Record<string, unknown>

const stable = (value: unknown): string => {
	if (value === null || typeof value !== 'object') return JSON.stringify(value) ?? 'undefined'
	if (Array.isArray(value)) return `[${value.map(stable).join(',')}]`

	return `{${Object.entries(value as Record<string, unknown>)
		.filter(([, entry]) => entry !== undefined)
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([key, entry]) => `${key}:${stable(entry)}`)
		.join(',')}}`
}

// distinguish carries values that identify the read but must not be sent as
// query params: paging handled by the caller, and filters applied after the
// response. Without them two such reads share a key and cancel each other.
export const keyed = (scope: string, options: ListOptions, distinguish?: unknown): ListOptions => ({
	...options,
	requestKey: `${scope}:${stable(options)}${distinguish === undefined ? '' : `:${stable(distinguish)}`}`,
})

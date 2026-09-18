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

export const keyed = (scope: string, options: ListOptions): ListOptions => ({
	...options,
	requestKey: `${scope}:${stable(options)}`,
})

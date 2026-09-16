import type { Sort, SortDirection } from '_/core/ports/sort'

export const asString = (value: unknown): string | undefined =>
	typeof value === 'string' && value !== '' ? value : undefined

export const asMember = <T extends string>(allowed: readonly T[], value: unknown): T | undefined =>
	allowed.includes(value as T) ? (value as T) : undefined

// Search params arrive as unknown from the URL, so each is narrowed to the
// domain's own union rather than trusted. An unparseable value is dropped
// instead of throwing, so a hand-edited URL degrades to a wider list.
export const asMembers = <T extends string>(allowed: readonly T[], value: unknown): readonly T[] | undefined => {
	const raw = Array.isArray(value) ? value : typeof value === 'string' ? value.split(',') : []
	const parsed = raw.map(entry => asMember(allowed, entry)).filter((entry): entry is T => entry !== undefined)

	return parsed.length > 0 ? parsed : undefined
}

export const asStrings = (value: unknown): readonly string[] | undefined => {
	const raw = Array.isArray(value) ? value : typeof value === 'string' ? value.split(',') : []
	const parsed = raw.map(asString).filter((entry): entry is string => entry !== undefined)

	return parsed.length > 0 ? parsed : undefined
}

const DIRECTIONS: readonly SortDirection[] = ['asc', 'desc']

// One param holds both halves as `field,direction`, so the pair can never
// arrive inconsistent. A navigate() call puts the object itself in the search
// state before it is ever serialised, so both shapes have to parse.
export const asSort = <T extends string>(allowed: readonly T[], value: unknown): Sort<T> | undefined => {
	const raw =
		typeof value === 'object' && value !== null
			? `${String((value as Sort<T>).field)},${String((value as Sort<T>).direction)}`
			: (asString(value) ?? '')

	const [rawField, rawDirection] = raw.split(',')
	const field = asMember(allowed, rawField)
	if (field === undefined) {
		return undefined
	}

	return { field, direction: asMember(DIRECTIONS, rawDirection) ?? 'desc' }
}

export const sortParam = <T extends string>(sort: Sort<T> | undefined): string | undefined =>
	sort ? `${sort.field},${sort.direction}` : undefined

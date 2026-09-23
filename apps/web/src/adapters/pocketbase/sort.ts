import { SORT_DIRECTION, type Sort } from '_/core/ports/sort'

// PocketBase names the timestamp columns `created` and `updated`, and takes a
// leading minus for descending.
const expression = <TField extends string>(sort: Sort<TField>): string =>
	`${sort.direction === SORT_DIRECTION.DESC ? '-' : ''}${sort.field}`

export const sortExpr = <TField extends string>(sort: Sort<TField> | undefined, fallback: string): string =>
	sort ? expression(sort) : fallback

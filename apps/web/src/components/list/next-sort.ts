import { SORT_DIRECTION, type Sort, type SortDirection } from '_/core/ports/sort'

const STARTS_DESCENDING: ReadonlySet<string> = new Set(['priority', 'size', 'created', 'updated'])

const firstDirection = (field: string): SortDirection =>
	STARTS_DESCENDING.has(field) ? SORT_DIRECTION.DESC : SORT_DIRECTION.ASC

export const nextSort = <TField extends string>(current: Sort<TField> | undefined, field: TField): Sort<TField> => {
	if (current?.field !== field) return { field, direction: firstDirection(field) }

	return { field, direction: current.direction === SORT_DIRECTION.DESC ? SORT_DIRECTION.ASC : SORT_DIRECTION.DESC }
}

export type Comparator = 'eq' | 'neq' | 'anyOf' | 'contains' | 'containsAll' | 'gte' | 'lte'

export type FilterValue = string | number | boolean | Date

export type Filterable = Record<string, FilterValue | readonly FilterValue[] | undefined>

export type Clause<T extends Filterable> = {
	[K in keyof T]-?: {
		readonly field: K
		readonly comparator: Comparator
		readonly value: T[K] | readonly NonNullable<T[K]>[] | undefined
	}
}[keyof T]

type Operator = {
	eq: '='
	neq: '!='
	contains: '~'
	gte: '>='
	lte: '<='
	anyOf: '='
	containsAll: '~'
}

type Joiner = {
	anyOf: '||'
	containsAll: '&&'
}

type Part<F extends string, C extends Comparator> = C extends keyof Joiner
	? `(${F} ${Operator[C]} {:pN} ${Joiner[C]} ...)`
	: `${F} ${Operator[C]} {:pN}`

type Expr<C> = C extends readonly [infer Head, ...infer Tail]
	? Head extends { field: infer F extends string; comparator: infer Cmp extends Comparator }
		? Tail extends readonly []
			? Part<F, Cmp>
			: `${Part<F, Cmp>} && ${Expr<Tail>}`
		: never
	: ''

type Bound<V> = V extends Date ? string : V extends readonly (infer I)[] ? Bound<I> : V

type BoundValues<C> = C extends readonly [infer Head, ...infer Tail]
	? Head extends { value: infer V }
		? Bound<NonNullable<V>> | BoundValues<Tail>
		: never
	: never

export type Filter<C extends readonly unknown[] = readonly unknown[]> = {
	readonly expr: Expr<C>
	readonly params: Readonly<Record<`p${number}`, BoundValues<C>>>
}

const operators = {
	eq: '=',
	neq: '!=',
	contains: '~',
	gte: '>=',
	lte: '<=',
	anyOf: '=',
	containsAll: '~',
} as const satisfies Operator

const joiners = {
	anyOf: '||',
	containsAll: '&&',
} as const satisfies Joiner

const isEmpty = (value: unknown): boolean =>
	value === undefined || value === null || value === '' || (Array.isArray(value) && value.length === 0)

const scalar = (value: unknown): string | number | boolean =>
	value instanceof Date ? value.toISOString() : (value as string | number | boolean)

export const filterFor =
	<T extends Filterable>() =>
	<const C extends readonly Clause<T>[]>(clauses: C): Filter<C> =>
		build(clauses)

const build = <const C extends readonly { field: unknown; comparator: Comparator; value: unknown }[]>(
	clauses: C,
): Filter<C> => {
	const parts: string[] = []
	const params: Record<string, string | number | boolean> = {}
	let next = 0

	const bind = (value: unknown): string => {
		const key = `p${next++}`
		params[key] = scalar(value)
		return `{:${key}}`
	}

	for (const { field, comparator, value } of clauses) {
		if (isEmpty(value)) continue

		const name = String(field)

		const operator = operators[comparator]

		if (comparator === 'anyOf' || comparator === 'containsAll') {
			const values: readonly unknown[] = Array.isArray(value) ? value : [value]
			const expanded = values.map(item => `${name} ${operator} ${bind(item)}`)
			parts.push(`(${expanded.join(` ${joiners[comparator]} `)})`)
			continue
		}

		parts.push(`${name} ${operator} ${bind(value)}`)
	}

	return {
		expr: parts.join(' && ') as Expr<C>,
		params: params as Filter<C>['params'],
	}
}

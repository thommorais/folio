const STORAGE_KEY = 'folio.list-filters'

const REMEMBERED = ['statuses', 'planStatuses', 'priority', 'tags', 'kinds', 'types', 'sort'] as const

type Search = Record<string, unknown>

const readAll = (): Record<string, Search> => {
	try {
		const parsed: unknown = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}')
		return typeof parsed === 'object' && parsed !== null ? (parsed as Record<string, Search>) : {}
	} catch {
		return {}
	}
}

const pick = (search: Search): Search =>
	Object.fromEntries(REMEMBERED.filter(key => search[key] !== undefined).map(key => [key, search[key]]))

const hasAny = (search: Search): boolean => Object.values(search).some(value => value !== undefined)

export const createListMemory = () => ({
	remember: (list: string, search: Search) => {
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...readAll(), [list]: pick(search) }))
		} catch {}
	},
	restore: (list: string, search: Search): Search | undefined => {
		if (hasAny(search)) return undefined

		const stored = readAll()[list]
		if (typeof stored !== 'object' || stored === null) return undefined

		const remembered = pick(stored)
		return hasAny(remembered) ? remembered : undefined
	},
})

export const listMemory = createListMemory()

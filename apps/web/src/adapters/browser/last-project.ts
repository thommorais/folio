const STORAGE_KEY = 'folio.last-project'

export type ProjectScope = {
	readonly client: string
	readonly domain: string
	readonly slug: string
}

const isScope = (value: unknown): value is ProjectScope => {
	if (typeof value !== 'object' || value === null) return false
	const { client, domain, slug } = value as Record<string, unknown>
	return typeof client === 'string' && typeof domain === 'string' && typeof slug === 'string'
}

const read = (): ProjectScope | undefined => {
	try {
		const parsed: unknown = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? 'null')
		return isScope(parsed) ? parsed : undefined
	} catch {
		return undefined
	}
}

export const createLastProject = () => {
	let resumed = false

	return {
		remember: (scope: ProjectScope) => {
			try {
				localStorage.setItem(STORAGE_KEY, JSON.stringify(scope))
			} catch {}
		},
		resume: (): ProjectScope | undefined => {
			if (resumed) return undefined
			resumed = true
			return read()
		},
		settle: () => {
			resumed = true
		},
	}
}

export const lastProject = createLastProject()

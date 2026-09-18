import { useParams } from '@tanstack/react-router'

type Scope = {
	readonly client: string
	readonly domain: string
}

// Every project route sits under /$client/$domain, so the pair is always in the
// URL. Reading it here keeps components from threading both through as props
// purely to rebuild a link.
export const useScope = (): Scope => {
	const params = useParams({ strict: false })

	return {
		client: typeof params.client === 'string' ? params.client : '',
		domain: typeof params.domain === 'string' ? params.domain : '',
	}
}

import type { Domain } from '_/core/domain/domain'
import { useAsyncState } from './realtime/use-async-state'
import { useContainer } from './container'

type DomainState =
	| { readonly status: 'loading' }
	| { readonly status: 'ready'; readonly domain: Domain }
	| { readonly status: 'failed'; readonly message: string }

export const useDomain = (client: string | undefined, slug: string | undefined): DomainState => {
	const { domains } = useContainer()

	const { state } = useAsyncState<Domain>(
		() => domains.get(client ?? '', slug ?? ''),
		[client, slug, domains],
		!client || !slug,
	)

	return state.status === 'ready' ? { status: 'ready', domain: state.data } : state
}

import type { Domain } from '_/core/domain/domain'
import { useAsyncState } from './realtime/use-async-state'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type DomainState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly domain: Domain }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useDomain = (client: string | undefined, slug: string | undefined): DomainState => {
	const { domains } = useContainer()

	const { state } = useAsyncState<Domain>(
		() => domains.get(client ?? '', slug ?? ''),
		[client, slug, domains],
		!client || !slug,
	)

	return state.status === Status.Ready ? { status: Status.Ready, domain: state.data } : state
}

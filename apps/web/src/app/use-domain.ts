import type { Domain } from '_/core/domain/domain'
import { useLiveRecord } from './realtime/use-live-record'
import { useContainer } from './container'
import { Status } from '_/lib/async-status'

type DomainState =
	| { readonly status: typeof Status.Idle }
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly domain: Domain }
	| { readonly status: typeof Status.Gone; readonly title: string }
	| { readonly status: typeof Status.Failed; readonly message: string }

export const useDomain = (client: string | undefined, slug: string | undefined): DomainState => {
	const { domains, connection } = useContainer()

	const state = useLiveRecord<Domain>({
		load: () => domains.get(client ?? '', slug ?? ''),
		subscribe: (id, onChange, onGone) => domains.subscribeToRecord(id, onChange, onGone),
		titleOf: domain => domain.name,
		connection,
		deps: [client, slug, domains],
		skip: !client || !slug,
	})

	return state.status === Status.Ready ? { status: Status.Ready, domain: state.data } : state
}

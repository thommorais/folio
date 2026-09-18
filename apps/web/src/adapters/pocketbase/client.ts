import { ENVS } from '_/envs'
import { TypedPocketBase } from '_/pocketbase-types'
import PocketBase from 'pocketbase'

const POCKETBASE_URL = ENVS.PUBLIC_API_URL

let browserClient: TypedPocketBase | undefined

const createPocketBaseClient = (): TypedPocketBase => {
	const client = new PocketBase(POCKETBASE_URL) as TypedPocketBase

	// Auto-cancellation aborts an in-flight request whenever another shares its
	// key, and two components reading the same data at once is ordinary rather
	// than a mistake: the issues list and the tree context behind it make the
	// same query by design. Aborting one of them turns a correct read into a
	// failure. A superseded read is already handled a level up, where
	// useAsyncState drops any response whose generation has moved on, so
	// nothing here depends on the transport cancelling it.
	client.autoCancellation(false)

	return client
}

const getPocketBaseClient = (): TypedPocketBase => {
	if (!browserClient) {
		browserClient = createPocketBaseClient()
	}

	return browserClient
}

export { getPocketBaseClient }

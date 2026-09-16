import { ENVS } from '_/envs'
import { TypedPocketBase } from '_/pocketbase-types'
import PocketBase from 'pocketbase'

const POCKETBASE_URL = ENVS.PUBLIC_API_URL

let browserClient: TypedPocketBase | undefined

const createPocketBaseClient = (): TypedPocketBase => {
	const pb = new PocketBase(POCKETBASE_URL) as TypedPocketBase

	if (import.meta.env.DEV) {
		pb.autoCancellation(false)
	}

	return pb
}

const getPocketBaseClient = (): TypedPocketBase => {
	// Client-side: reuse singleton
	if (!browserClient) {
		browserClient = createPocketBaseClient()
	}

	return browserClient
}

export { getPocketBaseClient }

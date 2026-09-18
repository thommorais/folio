import { ENVS } from '_/envs'
import { TypedPocketBase } from '_/pocketbase-types'
import PocketBase from 'pocketbase'

const POCKETBASE_URL = ENVS.PUBLIC_API_URL

let browserClient: TypedPocketBase | undefined

const createPocketBaseClient = (): TypedPocketBase => new PocketBase(POCKETBASE_URL) as TypedPocketBase

const getPocketBaseClient = (): TypedPocketBase => {
	if (!browserClient) {
		browserClient = createPocketBaseClient()
	}

	return browserClient
}

export { getPocketBaseClient }

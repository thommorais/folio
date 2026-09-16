import { ENVS } from '_/envs'
import type { ANY } from '_/types'
import type { StoreApi, UseBoundStore } from 'zustand'
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

type WithSelectors<S> = S extends { getState: () => infer T }
	? S & { [K in keyof T as `use${Capitalize<string & K>}`]: () => T[K] }
	: never

const createStoreSelectors = <S extends UseBoundStore<StoreApi<object>>>(_store: S) => {
	const store = _store as WithSelectors<typeof _store>
	for (const k of Object.keys(_store.getState())) {
		const hookName = `use${k.charAt(0).toUpperCase()}${k.slice(1)}`
		;(store as ANY)[hookName] = () => _store(s => s[k as keyof typeof s])
	}

	return store
}

type DevToolsType = typeof devtools

const conditionalDevtools: DevToolsType = ENVS.IS_DEV ? devtools : ((config => config) as DevToolsType)

export { conditionalDevtools, create, createStoreSelectors }

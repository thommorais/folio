import { z } from '_/lib/zod'

export const EMPTY = 'EMPTY' as const

const viteEnv = import.meta.env

const baseFlags = {
	IS_PROD: viteEnv.PROD,
	IS_DEV: viteEnv.DEV,
	IS_TEST: viteEnv.MODE === 'test',
} as const

const flags = {
	...baseFlags,
	IS_CLIENT: true,
	IS_SERVER: false,
} as const

type EnvsFromFlags<TFlags> = TFlags extends { IS_CLIENT: true } ? TFlags & ClientEnvs : TFlags

type Envs = EnvsFromFlags<typeof flags>

const createEnvs = (parsed: MergedSafeParseReturn): Envs => {
	if (parsed.success === false) {
		const message = 'Invalid environment variables'
		throw new Error(message)
	}

	const extendedEnvs = {
		...parsed.data,
		...flags,
	} as Envs

	const ENVS = new Proxy(extendedEnvs, {
		get(target, prop) {
			if (typeof prop !== 'string') {
				return undefined
			}

			if (prop in flags) {
				return Reflect.get(target, prop)
			}

			return Reflect.get(target, prop)
		},
	})

	return ENVS
}

const clientSchema = z.object({
	NODE_ENV: z.enum(['development', 'test', 'production']),
	PUBLIC_API_URL: z
		.union([z.url(), z.literal('')])
		.default('http://127.0.0.1:8090')
		.transform(value => (value === '' ? '/' : value)),
})

// Don't touch
// --------------------------

type ClientEnvs = z.infer<typeof clientSchema>
type EnvsKeys = keyof ClientEnvs
type PROCESS_ENV = Record<EnvsKeys, string | undefined>

type MergedSafeParseReturn = z.ZodSafeParseResult<ClientEnvs>

const parseEnvs = (processEnv: PROCESS_ENV, clientSchema: z.ZodSchema<ClientEnvs>): MergedSafeParseReturn => {
	const schema = clientSchema
	return schema.safeParse(processEnv)
}

const processEnv: PROCESS_ENV = {
	// Server-side env vars (unused in the SPA; kept for schema parity)
	NODE_ENV: viteEnv.MODE as 'development' | 'test' | 'production',
	PUBLIC_API_URL: viteEnv.VITE_API_URL,
}

const ENVS = createEnvs(parseEnvs(processEnv, clientSchema))

export { ENVS }

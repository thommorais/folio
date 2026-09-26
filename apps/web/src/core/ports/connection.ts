export type ConnectionPort = {
	readonly onReconnect: (listener: () => void) => () => void
	readonly retry: () => void
}

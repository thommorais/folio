export type ConnectionPort = {
	readonly onReconnect: (listener: () => void) => () => void
}

export type ConnectionPort = {
	readonly onReconnect: (listener: () => void) => () => void
	readonly retry: () => void
	readonly onStatusChange: (listener: (online: boolean) => void) => () => void
}

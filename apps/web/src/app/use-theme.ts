import { useCallback, useSyncExternalStore } from 'react'
import type { Appearance, Theme } from '_/core/ports/theme'
import { useContainer } from './container'

export const useTheme = (): {
	readonly theme: Theme
	readonly appearance: Appearance
	readonly setTheme: (theme: Theme) => void
} => {
	const { theme: port } = useContainer()

	const theme = useSyncExternalStore(port.subscribe, port.current, port.current)
	const appearance = useSyncExternalStore(port.subscribe, port.resolved, port.resolved)

	const setTheme = useCallback((next: Theme) => port.set(next), [port])

	return { theme, appearance, setTheme }
}

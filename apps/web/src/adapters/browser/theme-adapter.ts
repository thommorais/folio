import type { Appearance, Theme, ThemePort } from '_/core/ports/theme'
import { THEME, THEMES } from '_/core/ports/theme'

const STORAGE_KEY = 'folio.theme'

const isTheme = (value: unknown): value is Theme => THEMES.includes(value as Theme)

const read = (): Theme => {
	try {
		const stored = localStorage.getItem(STORAGE_KEY)
		return isTheme(stored) ? stored : THEME.SYSTEM
	} catch {
		return THEME.SYSTEM
	}
}

const systemQuery = () => window.matchMedia('(prefers-color-scheme: dark)')

const resolve = (theme: Theme): Appearance => {
	if (theme !== THEME.SYSTEM) return theme
	return systemQuery().matches ? THEME.DARK : THEME.LIGHT
}

export const applyAppearance = (appearance: Appearance): void => {
	document.documentElement.classList.toggle('dark', appearance === THEME.DARK)
	document.documentElement.style.colorScheme = appearance
}

export const createThemeAdapter = (): ThemePort => {
	const listeners = new Set<() => void>()
	let theme = read()
	let appearance = resolve(theme)

	const notify = () => {
		for (const listener of listeners) listener()
	}

	const sync = () => {
		appearance = resolve(theme)
		applyAppearance(appearance)
		notify()
	}

	return {
		current: () => theme,

		resolved: () => appearance,

		set: next => {
			theme = next
			try {
				localStorage.setItem(STORAGE_KEY, next)
			} catch {
				// A blocked store still themes this page; it just will not persist.
			}
			sync()
		},

		subscribe: listener => {
			listeners.add(listener)
			const media = systemQuery()
			const onSystemChange = () => {
				if (theme === THEME.SYSTEM) sync()
			}
			media.addEventListener('change', onSystemChange)

			return () => {
				listeners.delete(listener)
				media.removeEventListener('change', onSystemChange)
			}
		},
	}
}

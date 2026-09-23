export const THEME = {
	LIGHT: 'light',
	DARK: 'dark',
	SYSTEM: 'system',
} as const

export const THEMES = [THEME.LIGHT, THEME.DARK, THEME.SYSTEM] as const

export type Theme = (typeof THEMES)[number]

export type Appearance = typeof THEME.LIGHT | typeof THEME.DARK

export type ThemePort = {
	readonly current: () => Theme
	readonly resolved: () => Appearance
	readonly set: (theme: Theme) => void
	readonly subscribe: (listener: () => void) => () => void
}

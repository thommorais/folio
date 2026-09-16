export const THEMES = ['light', 'dark', 'system'] as const

export type Theme = (typeof THEMES)[number]

export type Appearance = 'light' | 'dark'

export type ThemePort = {
	readonly current: () => Theme
	readonly resolved: () => Appearance
	readonly set: (theme: Theme) => void
	readonly subscribe: (listener: () => void) => () => void
}

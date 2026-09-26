import { useEffect } from 'react'
import { OWNS_KEYS } from './type-to-search'

type Arrow = 'ArrowDown' | 'ArrowUp'

export const nextRow = (current: number, count: number, key: Arrow): number | undefined => {
	if (count === 0) return undefined
	if (current === -1) return key === 'ArrowDown' ? 0 : count - 1

	return key === 'ArrowDown' ? Math.min(current + 1, count - 1) : Math.max(current - 1, 0)
}

export const useRowNavigation = () => {
	useEffect(() => {
		const listener = (event: KeyboardEvent) => {
			if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return
			if (event.metaKey || event.ctrlKey || event.altKey || event.shiftKey) return
			if (event.target instanceof Element && event.target.closest(OWNS_KEYS)) return

			const rows = [...document.querySelectorAll<HTMLElement>('[data-row]')]
			const focused = document.activeElement instanceof Element ? document.activeElement.closest('[data-row]') : null
			const target = nextRow(focused ? rows.indexOf(focused as HTMLElement) : -1, rows.length, event.key)
			if (target === undefined) return

			event.preventDefault()
			rows[target]?.focus()
			rows[target]?.scrollIntoView({ block: 'nearest' })
		}

		document.addEventListener('keydown', listener)
		return () => document.removeEventListener('keydown', listener)
	}, [])
}

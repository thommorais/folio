import { type RefObject, useEffect, useEffectEvent } from 'react'

type KeyLike = {
	readonly key: string
	readonly metaKey: boolean
	readonly ctrlKey: boolean
	readonly altKey: boolean
	readonly target: { readonly closest: (selector: string) => unknown } | null
}

export const OWNS_KEYS = 'input, textarea, select, [contenteditable], [role="menu"], [role="dialog"], [role="listbox"]'

export const capturedKey = (event: KeyLike): { readonly insert: string } | undefined => {
	if (event.metaKey || event.ctrlKey || event.altKey) return undefined
	if (event.key.length !== 1 || event.key === ' ') return undefined
	if (event.target?.closest(OWNS_KEYS)) return undefined

	return { insert: event.key === '/' ? '' : event.key }
}

export const useTypeToSearch = (input: RefObject<HTMLInputElement | null>, onType: (text: string) => void) => {
	const type = useEffectEvent(onType)

	useEffect(() => {
		const listener = (event: KeyboardEvent) => {
			const captured = capturedKey({
				key: event.key,
				metaKey: event.metaKey,
				ctrlKey: event.ctrlKey,
				altKey: event.altKey,
				target: event.target instanceof Element ? event.target : null,
			})
			if (captured === undefined || input.current === null) return

			event.preventDefault()
			input.current.focus()
			if (captured.insert !== '') type(captured.insert)
		}

		document.addEventListener('keydown', listener)
		return () => document.removeEventListener('keydown', listener)
	}, [input])
}

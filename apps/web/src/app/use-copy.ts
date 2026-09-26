import { useCallback, useEffect, useState } from 'react'

const RESET_AFTER = 2000

export const useCopy = () => {
	const [copied, setCopied] = useState(false)

	useEffect(() => {
		if (!copied) return

		const timer = setTimeout(() => setCopied(false), RESET_AFTER)

		return () => clearTimeout(timer)
	}, [copied])

	const copy = useCallback(async (text: string) => {
		await navigator.clipboard.writeText(text)
		setCopied(true)
	}, [])

	return { copied, copy }
}

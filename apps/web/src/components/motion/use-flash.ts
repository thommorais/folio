import { useEffect, useState } from 'react'

const LIT_FOR = 1200

export const useFlash = (value: unknown): boolean => {
	const [seen, setSeen] = useState(value)
	const [lit, setLit] = useState(false)

	if (!Object.is(value, seen)) {
		setSeen(value)
		setLit(true)
	}

	useEffect(() => {
		if (!lit) return

		const timer = setTimeout(() => setLit(false), LIT_FOR)

		return () => clearTimeout(timer)
	}, [lit, seen])

	return lit
}

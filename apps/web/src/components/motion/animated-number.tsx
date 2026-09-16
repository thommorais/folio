import NumberFlow, { type Format } from '@number-flow/react'
import { useEffect, useRef } from 'react'

type AnimatedNumberProps = {
	readonly value: number
	readonly locale?: string
	readonly format?: Format
}

// Counts arrive as 0 before the first fetch resolves. Animating that would roll
// every tile up from zero on load, so the first real value is set without motion.
export const AnimatedNumber = ({ value, locale, format }: AnimatedNumberProps) => {
	const settled = useRef(false)

	useEffect(() => {
		if (value !== 0) settled.current = true
	}, [value])

	return <NumberFlow value={value} animated={settled.current} format={format} locales={locale ?? 'en'} willChange />
}

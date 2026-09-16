import { motion, useReducedMotion } from 'motion/react'
import type { ReactNode } from 'react'

const EASE = [0.16, 1, 0.3, 1] as const

// Rows past this point land after the eye has already moved on, so they all
// share the last delay rather than trailing a long list in sequence.
const MAX_STAGGERED = 8
const STEP = 0.03

type StaggerItemProps = {
	readonly index: number
	readonly children: ReactNode
	readonly className?: string
	// List rows live inside a <ul>, where a wrapping <div> would be invalid.
	readonly as?: 'div' | 'li'
}

export const StaggerItem = ({ index, children, className, as = 'div' }: StaggerItemProps) => {
	const reduced = useReducedMotion()
	const Tag = as === 'li' ? motion.li : motion.div

	if (reduced) {
		const Plain = as
		return <Plain className={className}>{children}</Plain>
	}

	return (
		<Tag
			className={className}
			initial={{ opacity: 0, y: 6 }}
			animate={{ opacity: 1, y: 0 }}
			transition={{ duration: 0.28, ease: EASE, delay: Math.min(index, MAX_STAGGERED) * STEP }}
		>
			{children}
		</Tag>
	)
}

type FadeInProps = {
	readonly children: ReactNode
	readonly className?: string
}

export const FadeIn = ({ children, className }: FadeInProps) => {
	const reduced = useReducedMotion()

	if (reduced) return <div className={className}>{children}</div>

	return (
		<motion.div
			className={className}
			initial={{ opacity: 0 }}
			animate={{ opacity: 1 }}
			transition={{ duration: 0.2, ease: EASE }}
		>
			{children}
		</motion.div>
	)
}

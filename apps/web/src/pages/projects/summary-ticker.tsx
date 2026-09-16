import { Link } from '@tanstack/react-router'
import { AnimatePresence, motion } from 'motion/react'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { Insight } from './insights'

const TICK_DURATION = 6000

const linkClass = 'border-dim/30 hover:text-foreground border-b border-dashed transition-colors'

const Line = ({ insight }: { readonly insight: Insight }) => (
	<>
		{insight.before}
		{insight.link &&
			(insight.to === undefined ? (
				insight.link
			) : (
				<Link to={insight.to} params={insight.params} className={linkClass}>
					{insight.link}
				</Link>
			))}
		{insight.after}
	</>
)

export const SummaryTicker = ({ insights }: { readonly insights: readonly Insight[] }) => {
	const [index, setIndex] = useState(0)
	const [direction, setDirection] = useState(1)
	const [fast, setFast] = useState(false)
	const [progress, setProgress] = useState(0)
	const hoveredRef = useRef(false)
	const startRef = useRef(Date.now())
	const elapsedOnPauseRef = useRef(0)

	const goTo = useCallback(
		(next: number) => {
			if (next === index) return

			setFast(true)
			setDirection(next > index ? 1 : -1)
			setIndex(next)
			setProgress(0)
			startRef.current = Date.now()
			elapsedOnPauseRef.current = 0
		},
		[index],
	)

	useEffect(() => {
		if (insights.length <= 1) return

		let raf: number

		const tick = () => {
			if (hoveredRef.current) {
				raf = requestAnimationFrame(tick)
				return
			}

			const elapsed = elapsedOnPauseRef.current + (Date.now() - startRef.current)
			const pct = Math.min(elapsed / TICK_DURATION, 1)
			setProgress(pct)

			if (pct >= 1) {
				setFast(false)
				setDirection(1)
				setIndex(prev => (prev + 1) % insights.length)
				setProgress(0)
				startRef.current = Date.now()
				elapsedOnPauseRef.current = 0
			}

			raf = requestAnimationFrame(tick)
		}

		raf = requestAnimationFrame(tick)

		return () => cancelAnimationFrame(raf)
	}, [insights.length])

	const current = insights[index % insights.length]

	if (current === undefined) return null

	if (insights.length <= 1) {
		return (
			<p className='text-dim mt-3 max-w-lg text-center text-sm leading-relaxed'>
				<Line insight={current} />
			</p>
		)
	}

	const slideY = (fast ? 8 : 14) * direction
	const duration = fast ? 0.12 : 0.32

	return (
		<div
			className='flex w-full max-w-lg flex-col items-center gap-3'
			onMouseEnter={() => {
				hoveredRef.current = true
				elapsedOnPauseRef.current += Date.now() - startRef.current
			}}
			onMouseLeave={() => {
				hoveredRef.current = false
				startRef.current = Date.now()
			}}
		>
			<div className='relative h-10 w-full overflow-hidden'>
				<AnimatePresence mode='wait' initial={false} custom={direction}>
					<motion.div
						key={current.key}
						className='absolute inset-0 flex items-center justify-center'
						initial={{ y: slideY, opacity: 0, filter: 'blur(4px)' }}
						animate={{ y: 0, opacity: 1, filter: 'blur(0px)' }}
						exit={{ y: -slideY, opacity: 0, filter: 'blur(4px)' }}
						transition={{ duration, ease: [0.16, 1, 0.3, 1] }}
					>
						<span className='text-dim text-center text-sm leading-relaxed'>
							<Line insight={current} />
						</span>
					</motion.div>
				</AnimatePresence>
			</div>

			<div className='flex items-center gap-1.5'>
				{insights.map((insight, position) => (
					<button
						key={insight.key}
						type='button'
						aria-label={`Show insight ${position + 1}`}
						className='group/bar relative -my-2 cursor-pointer py-2'
						onClick={() => goTo(position)}
						onMouseEnter={() => goTo(position)}
					>
						<div className='bg-foreground/10 relative h-[2px] w-4 overflow-hidden rounded-full'>
							<div
								className='bg-foreground/40 absolute inset-0 origin-left rounded-full group-hover/bar:scale-x-100!'
								style={{ transform: `scaleX(${position === index ? progress : 0})` }}
							/>
						</div>
					</button>
				))}
			</div>
		</div>
	)
}

import { useEffect, useState } from 'react'
import { useCurrentUser } from '_/app/use-session'

// Ported from midday's welcome-section, which resolves the zone from the user
// record first. Our session carries no timezone, so the browser's is all we have.
const greetingAt = (now: Date): string => {
	const hour = now.getHours()

	if (hour >= 5 && hour < 12) return 'Good morning'
	if (hour >= 12 && hour < 17) return 'Good afternoon'

	return 'Good evening'
}

const REFRESH = 5 * 60 * 1000

export const WelcomeGreeting = () => {
	const user = useCurrentUser()
	const [greeting, setGreeting] = useState(() => greetingAt(new Date()))

	useEffect(() => {
		const interval = setInterval(() => setGreeting(greetingAt(new Date())), REFRESH)

		return () => clearInterval(interval)
	}, [])

	const firstName = user?.name.trim().split(' ').at(0)

	return (
		<h1 className='text-center font-serif text-[38px] leading-tight'>
			{greeting}
			{firstName ? (
				<>
					, <span className='text-dim'>{firstName}</span>
				</>
			) : null}
		</h1>
	)
}

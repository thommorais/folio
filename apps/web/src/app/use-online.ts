import { useEffect, useState } from 'react'
import { useContainer } from './container'

const GRACE = 2000

export const useOnline = (): boolean => {
	const { connection } = useContainer()
	const [online, setOnline] = useState(true)

	useEffect(() => {
		let timer: ReturnType<typeof setTimeout> | undefined

		const stop = connection.onStatusChange(next => {
			clearTimeout(timer)

			if (next) {
				setOnline(true)
				return
			}

			timer = setTimeout(() => setOnline(false), GRACE)
		})

		return () => {
			clearTimeout(timer)
			stop()
		}
	}, [connection])

	return online
}

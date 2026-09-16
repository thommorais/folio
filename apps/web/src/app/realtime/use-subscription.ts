import type { Unsubscribe } from '_/core/ports/subscription'
import type { Result } from '_/lib/result'
import { useEffect } from 'react'

const noop = () => {}

type Open = () => Promise<Result<Unsubscribe>>

export const useSubscription = (open: Open, deps: readonly unknown[]): void => {
	useEffect(() => {
		const controller = new AbortController()
		let close: Unsubscribe | undefined

		void open()
			.then(result => {
				if (!result.success) return

				if (controller.signal.aborted) {
					void result.value().catch(noop)
					return
				}

				close = result.value
			})
			.catch(noop)

		return () => {
			controller.abort()
			void close?.().catch(noop)
		}
		// oxlint-disable-next-line react-hooks/exhaustive-deps
	}, deps)
}

import { toast } from '@thom/ui/toast'
import { registerServiceWorker } from '_/app/register-sw'
import { useEffect } from 'react'

// registerType is 'prompt', so a new worker waits rather than taking over
// mid-session: reloading under someone writing a plan would lose the form.
export const ReloadPrompt = () => {
	useEffect(() => {
		registerServiceWorker(({ update }) => {
			toast('A new version is ready.', {
				duration: Number.POSITIVE_INFINITY,
				action: {
					label: 'Reload',
					onClick: () => {
						void update()
					},
				},
			})
		})
	}, [])

	return null
}
